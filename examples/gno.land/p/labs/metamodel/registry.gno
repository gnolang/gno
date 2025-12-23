package metamodel

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// --- Types ---

type RegistryCallback = func(key string, value any) bool
type RegistryMatcher = func(value any) bool

type Registry struct {
	tree          *avl.Tree
	allowPrefixes []string
	renderOpts    map[string]any
	keywordIndex  *avl.Tree // maps keyword -> []string (keys)
	idIndex       *avl.Tree // maps id (int) -> string (key)
	sequence      int
}

// --- Constructor ---

func NewRegistry(allowedPrefixes []string, opts map[string]any) *Registry {
	chainId := std.ChainID()
	if _, ok := opts[chainId]; !ok {
		panic("render options must be set for the current chain: " + chainId)
	}
	renderOpts, ok := opts[chainId].(map[string]interface{})
	if !ok {
		panic("render options must be a map[string]interface{} for the current chain: " + chainId)
	}

	return &Registry{
		tree:          avl.NewTree(),
		allowPrefixes: allowedPrefixes,
		renderOpts:    renderOpts,
		keywordIndex:  avl.NewTree(),
		idIndex:       avl.NewTree(),
		sequence:      0,
	}
}

// --- Registration & Access Control ---

func (r *Registry) Register(key string, obj any, kw ...string) {
	r.assertAccess()
	r.register(key, obj, kw...)
}

func (r *Registry) register(name string, obj any, kw ...string) {
	key := sanitizeAnchorKey(name)
	if r.Exists(key) {
		panic("Key already exists: " + key)
	}
	kwords := keywords(append(kw, name))
	switch obj := obj.(type) {
	case *Model:
		r.addRecord(key, name, obj, kwords)
		r.indexKeywords(key, kwords)
		r.sequence++
	default:
		panic(ufmt.Sprintf("Invalid type for key %s: %T", key, obj))
	}
}

func (r *Registry) addRecord(key, name string, obj *Model, kwords []string) {
	rec := Record{
		Id:       r.sequence,
		Title:    name,
		Content:  func(...any) *Model { return obj },
		Keywords: kwords,
	}
	r.tree.Set(key, rec)
	r.idIndex.Set(strconv.Itoa(rec.Id), key) // convert int to string
}

func (r *Registry) KeyById(id int) (string, bool) {
	key, ok := r.idIndex.Get(strconv.Itoa(id)) // convert int to string
	if !ok {
		return "", false
	}
	return key.(string), true
}

func (r *Registry) indexKeywords(key string, kwords []string) {
	for _, word := range kwords {
		val, _ := r.keywordIndex.Get(word)
		var keys []string
		if val != nil {
			keys = val.([]string)
		}
		keys = append(keys, key)
		r.keywordIndex.Set(word, keys)
	}
}

func (r *Registry) Exists(key string) bool {
	_, ok := r.tree.Get(key)
	return ok
}

func (r *Registry) assertAccess() {
	if !r.hasAllowedPrefix() {
		panic("access denied: " + std.PreviousRealm().PkgPath() +
			" realm must match an allowed prefix:[" + strings.Join(r.allowPrefixes, ",") + "]")
	}
}

func (r *Registry) hasAllowedPrefix() bool {
	prevRealm := std.PreviousRealm()
	for _, callerPath := range r.allowPrefixes {
		if strings.HasPrefix(prevRealm.PkgPath(), callerPath) {
			return true
		}
	}
	return false
}

// --- Query & Indexing ---

func (r *Registry) KeysByKeyword(word string) []string {
	val, ok := r.keywordIndex.Get(strings.ToLower(word))
	if !ok {
		return nil
	}
	return val.([]string)
}

func (r *Registry) Iterate(callback RegistryCallback, matcher RegistryMatcher) {
	r.tree.IterateByOffset(0, r.tree.Size(), func(key string, value any) bool {
		if matcher(value) {
			return callback(key, value)
		}
		return false
	})
}

// --- Rendering & Content ---

func (r *Registry) Render(path string, cb ...func(string) string) string {
	q := parseQuery(path)
	if strings.HasPrefix(path, "?search=") {
		path = q["search"][0]
	}

	if strings.HasPrefix(path, "?") {
		if idx, ok := q["i"]; ok && len(idx) > 0 {
			return r.renderItem(idx[0], q)
		}
	}

	if path != "" {
		key := sanitizeAnchorKey(path)
		if _, ok := r.tree.Get(key); ok {
			return r.Content(key, path)
		}
		if c, ok := r.keywordIndex.Get(strings.ToLower(key)); ok {
			return r.renderSearchResults(key, c.([]string))
		}
	}

	return r.renderIndex(path, cb...)
}

func (r *Registry) getOptsString(key string) string {
	if r.renderOpts == nil {
		panic("render options not set")
	}
	if val, ok := r.renderOpts[key]; ok {
		return val.(string)
	}
	panic("render option not found: " + key)
}

func (r *Registry) renderSearchResults(key string, keys []string) string {
	const pageSize = 5
	var sb strings.Builder
	sb.WriteString("## Search Results for: " + key + "\n\n")
	total := len(keys)
	page, cleanKey := parsePageFromKey(key)
	start, end := getPageBounds(page, pageSize, total)
	sb.WriteString("Found " + ufmt.Sprintf("%d", total) + " results:\n\n")
	r.renderSearchItems(&sb, keys, start, end)
	r.renderSearchPagination(&sb, cleanKey, page, pageSize, total, end)
	sb.WriteString("\n[<- Home](" + r.getOptsString("BasePath") + ")\n")
	return sb.String()
}

func parsePageFromKey(key string) (page int, cleanKey string) {
	page = 0
	cleanKey = key
	if idx := strings.Index(key, "?p="); idx != -1 {
		pageStr := key[idx+3:]
		cleanKey = key[:idx]
		if i, err := strconv.Atoi(pageStr); err == nil {
			page = i
		}
	}
	return
}

func getPageBounds(page, pageSize, total int) (start, end int) {
	start = page * pageSize
	end = start + pageSize
	if end > total {
		end = total
	}
	return
}

func (r *Registry) renderSearchItems(sb *strings.Builder, keys []string, start, end int) {
	for i := start; i < end; i++ {
		k := keys[i]
		obj, ok := r.tree.Get(k)
		if !ok {
			continue
		}
		rec := obj.(Record)
		sb.WriteString("### " + rec.Title + "\n\n")
		sb.WriteString(r.thumbnail(k, obj) + "\n")
	}
}

func (r *Registry) renderSearchPagination(sb *strings.Builder, key string, page, pageSize, total, end int) {
	if total > pageSize {
		sb.WriteString("\n")
		if page > 0 {
			sb.WriteString("[Prev Page](" + r.getOptsString("BasePath") + ":" + key + "?p=" + ufmt.Sprintf("%d", page-1) + ") ")
		}
		sb.WriteString(ufmt.Sprintf("Page %d/%d", page+1, (total+pageSize)/pageSize+1))
		if end < total {
			sb.WriteString(" [Next Page](" + r.getOptsString("BasePath") + ":" + key + "?p=" + ufmt.Sprintf("%d", page+1) + ")")
		}
		sb.WriteString("\n")
	}
}

func (r *Registry) renderItem(indexStr string, u url.Values) string {
	index, _ := strconv.Atoi(indexStr)
	i := index % r.tree.Size()
	key, obj := r.tree.GetByIndex(i)
	rec := obj.(Record)
	var sb strings.Builder
	sb.WriteString("id: " + ufmt.Sprintf("%d", rec.Id) + "\n\n")
	sb.WriteString("### " + upcaseWords(rec.Title) + "\n\n")
	sb.WriteString("**[" + key + "🔗](" + r.getOptsString("BasePath") + ":" + key + ")**\n\n")
	sb.WriteString("**Keywords:** " + strings.Join(rec.Keywords, ", ") + "\n\n")
	sb.WriteString(r.renderNavigation(i, u))
	sb.WriteString(r.Content(key, ""))
	sb.WriteString(r.renderNavigation(i, u))
	sb.WriteString("\n[<- Home](" + r.getOptsString("BasePath") + ")\n")
	return sb.String()
}

func (r *Registry) renderNavigation(i int, u url.Values) string {
	var sb strings.Builder
	prev := (i - 1 + r.tree.Size()) % r.tree.Size()
	next := (i + 1) % r.tree.Size()
	sb.WriteString("[<-prev](?i=" + ufmt.Sprintf("%d", prev) + ") ")
	sb.WriteString(ufmt.Sprintf("%d/%d", i+1, r.tree.Size()))
	sb.WriteString(" [next->](?i=" + ufmt.Sprintf("%d", next) + ")\n\n")
	return sb.String()
}

func (r *Registry) SearchForm() string {
	return "View all models on the [Metamodel List ->](" +
		r.getOptsString("BasePath") + "?p=0" +
		")\n\n" + searchForm
}

var searchForm = `
<gno-columns>
<gno-form>
    <gno-input name="search" placeholder="keyword" />
</gno-form>
|||

**Search Models**:

Enter a keyword

or

[Review all models ->](?p=0)

</gno-columns>
`

func (r *Registry) renderIndex(path string, cb ...func(string) string) string {
	q := parseQuery(path)
	if page, ok := q["p"]; ok && len(page) > 0 {
		return r.renderList(path)
	}
	var sb strings.Builder
	for _, content := range cb {
		if content != nil {
			sb.WriteString(content(path) + "\n\n")
		}
	}
	sb.WriteString("### Index\n\n")
	sb.WriteString(searchForm + "\n\n")

	return sb.String()
}

func (r *Registry) renderList(path string) string {
	const pageSize = 5
	total := r.tree.Size()
	page := 0
	if u := parseQuery(path); u != nil {
		if idx, ok := u["p"]; ok && len(idx) > 0 {
			if i, err := strconv.Atoi(idx[0]); err == nil {
				page = i
			}
		}
	}
	sb := r.renderListItems(page, pageSize, total)
	nav := r.renderListNav(page, pageSize, total)
	sb.WriteString(nav)
	return "## Metamodel Index\n\n" +
		nav +
		sb.String() +
		"\n[<- Home](" + r.getOptsString("BasePath") + ")\n"
}

func (r *Registry) renderListItems(page, pageSize, total int) *strings.Builder {
	sb := &strings.Builder{}
	start := page * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	for i := start; i < end; i++ {
		key, obj := r.tree.GetByIndex(i)
		sb.WriteString("### [" + upcaseWords(key) + "](?i=" + ufmt.Sprintf("%d", i) + ")\n")
		sb.WriteString(r.thumbnail(key, obj))
	}
	return sb
}

func (r *Registry) renderListNav(page, pageSize, total int) string {
	var nav strings.Builder
	if total > pageSize {
		nav.WriteString("\n")
		if page > 0 {
			nav.WriteString("[Prev Page](?p=" + ufmt.Sprintf("%d", page-1) + ") ")
		}
		nav.WriteString(ufmt.Sprintf("Page %d/%d", page+1, (total+pageSize)/pageSize+1))
		if (page+1)*pageSize < total {
			nav.WriteString(" [Next Page](?p=" + ufmt.Sprintf("%d", page+1) + ")")
		}
		nav.WriteString("\n")
	}
	return nav.String()
}

func (r *Registry) Content(key, path string) string {
	key = sanitizeAnchorKey(key)
	val, ok := r.tree.Get(key)
	if !ok {
		panic("Key not found: " + key)
	}
	rec := val.(Record)
	model := rec.Content.(func(...any) *Model)()
	if binding, ok := model.Binding.(func(path string) string); ok {
		return binding(path)
	}
	return model.Markdown()
}

func (r *Registry) Preview(key, path string) string {
	key = sanitizeAnchorKey(key)
	val, ok := r.tree.Get(key)
	if !ok {
		panic("Key not found: " + key)
	}
	rec := val.(Record)
	model := rec.Content.(func(...any) *Model)()
	lineCount := 3
	var content string
	if binding, ok := model.Binding.(func(string) string); ok {
		content = binding(path)
		lines := strings.Split(content, "\n")
		if len(lines) > lineCount {
			return strings.Join(lines[:lineCount], "\n") + "\n"
		}
	} else {
		content = model.Markdown()
	}
	return markdownPreview(content)
}

func (r *Registry) thumbnail(key string, obj any) string {
	rec := obj.(Record)
	var sb strings.Builder
	sb.WriteString("**Keywords:** " + strings.Join(rec.Keywords, ", ") + "\n\n")
	model := rec.Content.(func(...any) *Model)()
	sb.WriteString("[![" + key + "](" + model.ThumbnailDataUrl() + ")](" + r.getOptsString("BasePath") + ":" + key + ")\n")
	sb.WriteString(r.Preview(key, "") + "\n")
	return sb.String()
}

// --- Helpers ---

func keywords(kw []string) (out []string) {
	seen := make(map[string]struct{})
	for _, k := range kw {
		for _, word := range strings.Split(k, " ") {
			word = strings.ToLower(word)
			if len(word) == 0 {
				continue
			}
			if _, exists := seen[word]; !exists {
				seen[word] = struct{}{}
				out = append(out, word)
			}
		}
	}
	sort.Strings(out)
	return out
}

func parseQuery(path string) url.Values {
	if len(path) > 0 && path[0] == '?' {
		u, err := url.Parse(std.CurrentRealm().PkgPath() + path)
		if err == nil {
			return u.Query()
		}
	}
	return url.Values{}
}

func upcaseWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func markdownPreview(content string) string {
	regexImages := `!\[.*?\]\(.*?\)`
	content = strings.ReplaceAll(content, regexImages, "")
	return content
}

func sanitizeAnchorKey(key string) string {
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, " ", "-")
	var sb strings.Builder
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func EscapeHtml(s string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#39;",
	}
	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}
	return s
}

func SubmitButton(label, path string, fontSize, minWidth int) string {
	charWidth := int(0.6 * float64(fontSize))
	padding := 40
	h := 2 * fontSize
	w := len(label)*charWidth + padding
	if w < minWidth {
		w = minWidth
	}

	svgButton := `<svg xmlns="http://www.w3.org/2000/svg" width="` + strconv.Itoa(w) + `" height="` + strconv.Itoa(h) + `">
<foreignObject x="16" y="-5" width="` + strconv.Itoa(w) + `" height="` + strconv.Itoa(h) + `">
  <body xmlns="http://www.w3.org/1999/xhtml">
    <button style="padding-left: 20px; font-size:` + strconv.Itoa(fontSize) + `px">
      ` + EscapeHtml(label) + `
    </button>
  </body>
</foreignObject>
</svg>`

	dataUrl := "data:image/svg+xml;utf8," + url.PathEscape(svgButton)
	return "[![" + label + "](" + dataUrl + ")](" + path + ")"
}

type Record struct {
	Id       int
	Title    string
	Content  any
	Keywords []string
}

func (record Record) String() string {
	var sb strings.Builder
	sb.WriteString("Record{Content: ")
	if record.Content != nil {
		sb.WriteString("any")
	} else {
		sb.WriteString("nil")
	}
	sb.WriteString(", Keywords: [")
	for i, keyword := range record.Keywords {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(keyword)
	}
	sb.WriteString("]}")
	return sb.String()
}

package releases

type release struct {
	// manual
	version string
	url     string
	notes   string

	// internal
	isLatest  bool
	changelog *changelog
}

func (r *release) URL() string     { return r.url }
func (r *release) Version() string { return r.version }
func (r *release) Notes() string   { return r.notes }
func (r *release) IsLatest() bool  { return r.isLatest }

func (r *release) Title() string {
	output := r.changelog.name + " " + r.version
	if r.isLatest {
		output += " (latest)"
	}
	return output
}

func (r *release) Link() string {
	return "[" + r.Title() + "](" + r.url + ")"
}

func (r *release) Render() string {
	output := ""
	output += "## " + r.Link() + "\n\n"
	if r.notes != "" {
		output += r.notes + "\n\n"
	}
	return output
}

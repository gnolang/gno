package feed

// MarshalJSON marshals the components of the feed that are needed for
// an agent to execute tasks and send values for ingestion.
func MarshalFeedJSON(feed Feed) ([]byte, error) {
    buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	w.Write([]byte(
		`{"id":"` + feed.id +
			`","type":"` + ufmt.Sprintf("%d", int(feed.Type())) +
			`","value_type":"` + feed.valueDataType +
			`","tasks":[`),
	)

	first := true
	for _, task := range feed.tasks {
		if !first {
			w.WriteString(",")
		}

		taskJSON, err := task.MarshalJSON()
		if err != nil {
			return nil, err
		}

		w.Write(taskJSON)
		first = false
	}

	w.Write([]byte("]}"))
	w.Flush()

	return buf.Bytes(), nil
}
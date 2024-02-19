package json

// ParsePath takes a JSONPath string and returns a slice of strings representing the path segments.
func ParsePath(path string) ([]string, error) {
    buf := NewBuffer([]byte(path))
    result := make([]string, 0)

    for {
        b, err := buf.current()
        if err != nil {
            break
        }

        switch {
        case b == DollarToken || b == AtToken:
            result = append(result, string(b))
			buf.step()

        case b == DotToken:
            buf.step()

            if next, _ := buf.current(); next == DotToken {
                buf.step()
                result = append(result, "..")

				extractNextSegment(buf, &result)
			} else {
				extractNextSegment(buf, &result)
            }

        case b == SquareOpenToken:
            start := buf.index
            buf.step()

            for {
                if buf.index >= buf.length || buf.data[buf.index] == SquareCloseToken {
                    break
                }

                buf.step()
            }

            if buf.index >= buf.length {
                return nil, newJSONError(EofError, "unexpected end of path")
            }

            segment := string(buf.sliceFromIndices(start+1, buf.index))
            result = append(result, segment)

            buf.step()

        default:
            buf.step()
        }
    }

    return result, nil
}

// extractNextSegment extracts the segment from the current index 
// to the next significant character and adds it to the resulting slice.
func extractNextSegment(buf *buffer, result *[]string) {
    start := buf.index
    buf.skipToNextSignificantToken()
    
    if buf.index <= start {
        return
    }
    
    segment := string(buf.sliceFromIndices(start, buf.index))
    if segment != "" {
        *result = append(*result, segment)
    }
}

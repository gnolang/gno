package home

func Render(path string) string {
	out := `<gno-form>

	<gno-markdown>
# Newsletter subscription
**Do you agree** to receive updates and announcements by email?
	</gno-markdown>	

	<gno-input name="newsletter" type="checkbox" value="subscribe" placeholder="Subscribe to newsletter" description="Receive updates about new features" />
	</gno-form>

`
	return out
}

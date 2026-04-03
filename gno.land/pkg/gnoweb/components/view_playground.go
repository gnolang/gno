package components

const PlaygroundViewType ViewType = "playground-view"

type PlaygroundData struct {
	// InitialCode is pre-filled code (e.g. from fork)
	InitialCode string
	// ForkFrom is the package path this was forked from
	ForkFrom string
	// Remote is the RPC endpoint
	Remote string
	// ChainId is the current chain ID
	ChainId string
	// Domain is the node domain
	Domain string
}

type playgroundViewParams struct {
	PlaygroundData
	Article ArticleData
}

func PlaygroundView(data PlaygroundData) *View {
	content := NewTemplateComponent("ui/playground_editor", data)
	viewData := playgroundViewParams{
		PlaygroundData: data,
		Article: ArticleData{
			ComponentContent: content,
			Classes:          "c-playground-view",
		},
	}

	return NewTemplateView(PlaygroundViewType, "renderPlayground", viewData)
}

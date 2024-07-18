package main

import (
	_ "embed"
)

// //go:embed assets/*.txt
// var banners embed.FS

// func getBannerFrames(prefix string) []string {
// 	files, err := banners.ReadDir("assets")
// 	if err != nil {
// 		panic(err)
// 	}

// 	frames := []string{}
// 	for _, file := range files {
// 		if file.IsDir() {
// 			continue
// 		}

// 		if !strings.HasPrefix(file.Name(), prefix) {
// 			continue
// 		}

// 		frame, err := banners.ReadFile(filepath.Join("assets", file.Name()))
// 		if err != nil {
// 			panic(err)
// 		}

// 		frames = append(frames, string(frame))
// 	}
// 	return frames
// }

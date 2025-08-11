package std

func X_deleteSliceIndex(data []string, index int) {
	data = append(data[:index], data[index+1:]...)
}

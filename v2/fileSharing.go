package xlsx

type FileSharing struct {
	ReadOnlyRecommended bool
}

func (instance *FileSharing) makeXLSXFileSharing() *xlsxFileSharing {
	return &xlsxFileSharing{
		ReadOnlyRecommended: instance.ReadOnlyRecommended,
	}
}

func (instance *FileSharing) fromXLSXFileSharing(in *xlsxFileSharing) error {
	instance.ReadOnlyRecommended = in.ReadOnlyRecommended
	return nil
}

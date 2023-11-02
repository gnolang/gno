package fs

import (
	"errors"
	"path"
)

type SubFS interface {
	FS
	Sub(dir string) (FS, error)
}

func Sub(fsys FS, dir string) (FS, error) {
	if !ValidPath(dir) {
		return nil, &PathError{Op: "sub", Path: dir, Err: errors.New("invalid name")}
	}

	if dir == "." {
		return fsys, nil
	}

	if fsys, ok := fsys.(SubFS); ok {
		return fsys.Sub(dir)
	}

	return &subFS{fsys, dir}, nil
}

type subFS struct {
	fsys FS
	dir  string
}

// fullName maps name to the fully-qualified name dir/name.
func (f *subFS) fullName(op string, name string) (string, error) {
	if !ValidPath(name) {
		return "", &PathError{Op: op, Path: name, Err: errors.New("invalid name")}
	}

	return path.Join(f.dir, name), nil
}

// shorten maps name, which should start with f.dir, back to the suffix after f.dir.
func (f *subFS) shorten(name string) (rel string, ok bool) {
	if name == f.dir {
		return ".", true
	}
	if len(name) >= len(f.dir)+2 && name[len(f.dir)] == '/' && name[:len(f.dir)] == f.dir {
		return name[len(f.dir)+1:], true
	}
	return "", false
}

// fixErr shortens any reported names in PathErrors by tripping f.dir.
func (f *subFS) fixErr(err error) error {
	if e, ok := err.(*PathError); ok {
		if short, ok := f.shorten(e.Path); ok {
			e.Path = short
		}
	}

	return err
}

func (f *subFS) Open(name string) (File, error) {
	full, err := f.fullName("open", name)
	if err != nil {
		return nil, err
	}

	file, err := f.fsys.Open(full)
	return file, f.fixErr(err)
}

func (f *subFS) ReadDir(name string) ([]DirEntry, error) {
	full, err := f.fullName("read", name)
	if err != nil {
		return nil, err
	}

	dir, err := ReadDir(f.fsys, full)

	return dir, f.fixErr(err)
}

func (f *subFS) ReadFile(name string) ([]byte, error) {
	full, err := f.fullName("read", name)
	if err != nil {
		return nil, err
	}

	data, err := ReadFile(f.fsys, full)
	return data, f.fixErr(err)
}

func (f *subFS) Glob(pattern string) ([]string, error) {
	if _, err := path.Match(pattern, ""); err != nil {
		return nil, err
	}

	if pattern == "." {
		return []string{"."}, nil
	}

	full := f.dir + "/" + pattern
	list, err := Glob(f.fsys, full)

	for i, name := range list {
		name, ok := f.shorten(name)
		if !ok {
			return nil, errors.New("invalid result from inner fsys Glob: " + name + " not in " + f.dir) // can't use fmt in this package
		}

		list[i] = name
	}

	return list, f.fixErr(err)
}

func (f *subFS) Sub(dir string) (FS, error) {
	if dir == "." {
		return f, nil
	}

	full, err := f.fullName("sub", dir)
	if err != nil {
		return nil, err
	}

	return &subFS{
		fsys: f.fsys,
		dir:  full,
	}, nil
}

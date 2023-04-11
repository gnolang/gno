package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

const (
	qFile = "vm/qfile"
)

var ErrPackageNotFound = errors.New("not found")

type TMClient struct {
	c *client.HTTP
}

func NewTMClient(addr string) *TMClient {
	return &TMClient{
		c: client.NewHTTP(addr, "/websocket"),
	}
}

func (tc *TMClient) GetGnoZip(module, version string, w io.Writer) error {
	paths, err := tc.getfilepathsRec(module)
	if err != nil {
		return err
	}

	zw := zip.NewWriter(w)

	for _, p := range paths {
		q, err := tc.c.ABCIQuery(qFile, []byte(p))
		if err != nil {
			return fmt.Errorf("error querying tendermint network: %w", err)
		}

		if q.Response.Error != nil {
			return fmt.Errorf("query response error from tendermint network: %s", q.Response.Error.Error())
		}

		buf := bytes.NewBuffer(q.Response.Data)
		np := strings.ReplaceAll(p, module, fmt.Sprintf("%s@%s", module, version))

		zfw, err := zw.Create(np)
		if err != nil {
			return fmt.Errorf("error adding file with path %q to zip: %w", np, err)
		}

		_, err = buf.WriteTo(zfw)
		if err != nil {
			return fmt.Errorf("error adding file with path %q to zip: %w", np, err)
		}
	}

	return zw.Close()
}

func (tc *TMClient) GetGoZip(gnoModule, goModule, version string, w io.Writer) error {
	paths, err := tc.getfilepathsRec(gnoModule)
	if err != nil {
		return err
	}

	zw := zip.NewWriter(w)

	for _, p := range paths {
		q, err := tc.c.ABCIQuery(qFile, []byte(p))
		if err != nil {
			return fmt.Errorf("error querying tendermint network: %w", err)
		}

		if q.Response.Error != nil {
			return fmt.Errorf("query response error from tendermint network: %s", q.Response.Error.Error())
		}
		tfn, tags := gnolang.GetPrecompileFilenameAndTags(p)
		res, err := gnolang.Precompile(string(q.Response.Data), tags, tfn)
		if err != nil {
			return fmt.Errorf("error precompiling file %q: %w", p, err)
		}

		buf := bytes.NewBuffer([]byte(res.Translated))
		replacer := strings.NewReplacer(
			path.Base(p), tfn,
			gnoModule, fmt.Sprintf("%s@%s", goModule, version),
		)
		np := replacer.Replace(p)

		zfw, err := zw.Create(np)
		if err != nil {
			return fmt.Errorf("error adding file with path %q to zip: %w", np, err)
		}

		_, err = buf.WriteTo(zfw)
		if err != nil {
			return fmt.Errorf("error adding file with path %q to zip: %w", np, err)
		}
	}

	return zw.Close()
}

func (tc *TMClient) getfilepathsRec(p string) ([]string, error) {
	var out []string

	// this is the only way to know right now if the obtained data is a file or a folder
	ext := path.Ext(p)
	if ext != "" {
		// it is a file, adding into the list and returning
		out = append(out, p)
		return out, nil
	}

	q, err := tc.c.ABCIQuery(qFile, []byte(p))
	if err != nil {
		return nil, fmt.Errorf("error getting info from file %q: %w", p, err)
	}

	if q.Response.Error != nil {
		return nil, fmt.Errorf("query response error from tendermint network: %s: %w", q.Response.Error.Error(), ErrPackageNotFound)
	}

	files := strings.Split(string(q.Response.Data), "\n")
	for _, f := range files {
		o, err := tc.getfilepathsRec(path.Join(p, f))
		if err != nil {
			return nil, err
		}

		out = append(out, o...)
	}

	return out, nil
}

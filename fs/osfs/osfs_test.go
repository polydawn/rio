package osfs

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/tests"
	"go.polydawn.net/rio/testutil"
)

func TestAll(t *testing.T) {
	for _, spec := range []func(*testing.T, fs.FS){
		tests.CheckMkdirLstatRoundtrip,
		tests.CheckDeepMkdirError,
		tests.CheckMklinkLstatRoundtrip,
	} {
		t.Run(fnname(spec), func(t *testing.T) {
			testutil.WithTmpdir(func(tmpDir fs.AbsolutePath) {
				tfs := New(tmpDir)
				boxPath := fs.MustRelPath("sandbox")
				tfs.Mkdir(boxPath, 0755)
				afs := New(tmpDir.Join(boxPath))

				spec(t, afs)
			})
		})
	}
}

func fnname(fn interface{}) string {
	fullname := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return fullname[strings.LastIndex(fullname, ".")+1:]
}

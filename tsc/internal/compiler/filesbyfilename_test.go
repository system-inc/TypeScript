package compiler

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

// GetSourceFile by a name the program holds its file under answers from an index, with no path computed, and
// the answer is the one the path lookup gives for every name: the held names, the same names cased otherwise
// on a case-insensitive file system, and a name the program does not hold. On a case-insensitive file system
// computing the path lowercases the name, so the held-name lookup is the one that must not allocate.
//
// Not parallel: testing.AllocsPerRun refuses to run in a parallel test.
func TestGetSourceFileByAHeldNameMatchesThePathLookup(t *testing.T) {
	files := map[string]any{
		"/src/tsconfig.json": `{"compilerOptions": {"noLib": true, "noEmit": true, "module": "esnext", "moduleResolution": "bundler"}, "include": ["*.ts"]}`,
		"/src/Geometry.ts":   "export const area = 1;\n",
		"/src/main.ts":       "import { area } from './Geometry';\nexport const total = area;\n",
	}
	for _, caseSensitive := range []bool{true, false} {
		host := NewCompilerHost("/", vfstest.FromMap(files, caseSensitive), "", nil, nil, nil)
		config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile("/src/tsconfig.json", nil, nil, host, nil)
		assert.Equal(t, len(diagnostics), 0)
		program := NewProgram(ProgramOptions{Config: config, Host: host})

		names := []string{"/src/missing.ts", "/src/GEOMETRY.TS", "/src/Main.ts"}
		for _, file := range program.GetSourceFiles() {
			names = append(names, file.FileName())
		}
		for _, name := range names {
			assert.Equal(t, program.GetSourceFile(name), program.GetSourceFileByPath(program.toPath(name)), "%s (case-sensitive %v)", name, caseSensitive)
		}
		geometry := program.GetSourceFile("/src/Geometry.ts")
		assert.Assert(t, geometry != nil)
		if allocations := testing.AllocsPerRun(100, func() { program.GetSourceFile("/src/Geometry.ts") }); allocations != 0 {
			t.Errorf("a lookup by a held name allocated %v times (case-sensitive %v)", allocations, caseSensitive)
		}
	}
}

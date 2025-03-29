package gs_conf

import (
	"os"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util/assert"
	"github.com/go-spring/spring-core/util/sysconf"
)

func clean() {
	os.Args = nil
	os.Clearenv()
	sysconf.Clear()
}

func TestAppConfig(t *testing.T) {
	clean()

	t.Run("resolve error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "${a}")
		_, err := NewAppConfig().Refresh()
		assert.Error(t, err, `resolve string "\${a}" error << property a not exist`)
	})

	t.Run("resolve error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-REMOTE_DIR", "${a}")
		_, err := NewAppConfig().Refresh()
		assert.Error(t, err, `resolve string "\${a}" error << property a not exist`)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "./testdata/conf")
		_ = os.Setenv("GS_SPRING_APP_CONFIG-REMOTE_DIR", "./testdata/conf/remote")
		p, err := NewAppConfig().Refresh()
		assert.Nil(t, err)
		assert.Equal(t, p.Data(), map[string]string{
			"spring.app.config-local.dir":  "./testdata/conf",
			"spring.app.config-remote.dir": "./testdata/conf/remote",
			"spring.app.name":              "remote",
			"http.server.addr":             "0.0.0.0:8080",
		})
	})

	t.Run("merge error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_A", "a")
		_ = os.Setenv("GS_A_B", "a.b")
		_, err := NewAppConfig().Refresh()
		assert.Error(t, err, "property 'a' is a value but 'a.b' wants other type")
	})

	t.Run("merge error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "./testdata/conf")
		_ = sysconf.Set("http.server[0].addr", "0.0.0.0:8080")
		_, err := NewAppConfig().Refresh()
		assert.Error(t, err, "property 'http.server' is an array but 'http.server.addr' wants other type")
	})
}

func TestBootConfig(t *testing.T) {
	clean()

	t.Run("resolve error", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "${a}")
		_, err := NewBootConfig().Refresh()
		assert.Error(t, err, `resolve string "\${a}" error << property a not exist`)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "./testdata/conf")
		p, err := NewBootConfig().Refresh()
		assert.Nil(t, err)
		assert.Equal(t, p.Data(), map[string]string{
			"spring.app.config-local.dir": "./testdata/conf",
			"spring.app.name":             "test",
			"http.server.addr":            "0.0.0.0:8080",
		})
	})

	t.Run("merge error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_A", "a")
		_ = os.Setenv("GS_A_B", "a.b")
		_, err := NewBootConfig().Refresh()
		assert.Error(t, err, "property 'a' is a value but 'a.b' wants other type")
	})

	t.Run("merge error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "./testdata/conf")
		_ = sysconf.Set("http.server[0].addr", "0.0.0.0:8080")
		_, err := NewBootConfig().Refresh()
		assert.Error(t, err, "property 'http.server' is an array but 'http.server.addr' wants other type")
	})
}

func TestPropertySources(t *testing.T) {
	clean()

	t.Run("add dir error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddDir("non_existent_dir")
		assert.Equal(t, 1, len(ps.extraDirs))
	})

	t.Run("add dir error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		assert.Panic(t, func() {
			ps.AddDir("./testdata/conf/app.properties")
		}, "should be a directory")
	})

	t.Run("add file error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("non_existent_file")
		assert.Equal(t, 1, len(ps.extraFiles))
	})

	t.Run("add file error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		assert.Panic(t, func() {
			ps.AddFile("./testdata/conf")
		}, "should be a file")
	})

	t.Run("reset", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("./testdata/conf/app.properties")
		assert.Equal(t, 1, len(ps.extraFiles))
		ps.AddDir("./testdata/conf")
		assert.Equal(t, 1, len(ps.extraDirs))
		ps.Reset()
		assert.Equal(t, 0, len(ps.extraFiles))
		assert.Equal(t, 0, len(ps.extraDirs))
	})

	t.Run("getDefaultDir - 1", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		dir, err := ps.getDefaultDir(conf.Map(nil))
		assert.Nil(t, err)
		assert.Equal(t, "./conf", dir)
	})

	t.Run("getDefaultDir - 2", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeRemote, "app")
		dir, err := ps.getDefaultDir(conf.Map(nil))
		assert.Nil(t, err)
		assert.Equal(t, "./conf/remote", dir)
	})

	t.Run("getFiles - 1", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		files, err := ps.getFiles("./conf", conf.Map(nil))
		assert.Nil(t, err)
		assert.Equal(t, files, []string{
			"./conf/app.properties",
			"./conf/app.yaml",
			"./conf/app.toml",
			"./conf/app.json",
		})
	})

	t.Run("getFiles - 2", func(t *testing.T) {
		t.Cleanup(clean)
		p := conf.Map(map[string]interface{}{
			"spring.profiles.active": "dev,test",
		})
		ps := NewPropertySources(ConfigTypeLocal, "app")
		files, err := ps.getFiles("./conf", p)
		assert.Nil(t, err)
		assert.Equal(t, files, []string{
			"./conf/app.properties",
			"./conf/app.yaml",
			"./conf/app.toml",
			"./conf/app.json",
			"./conf/app-dev.properties",
			"./conf/app-dev.yaml",
			"./conf/app-dev.toml",
			"./conf/app-dev.json",
			"./conf/app-test.properties",
			"./conf/app-test.yaml",
			"./conf/app-test.toml",
			"./conf/app-test.json",
		})
	})

	t.Run("loadFiles", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("./testdata/conf/app.properties")
		files, err := ps.loadFiles(conf.Map(nil))
		assert.Nil(t, err)
		assert.Equal(t, 1, len(files))
	})

	t.Run("loadFiles - getDefaultDir error", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources("invalid", "app")
		_, err := ps.loadFiles(conf.Map(nil))
		assert.Error(t, err, "unknown config type: invalid")
	})

	t.Run("loadFiles - getFiles error", func(t *testing.T) {
		t.Cleanup(clean)
		p := conf.Map(map[string]interface{}{
			"spring.profiles.active": "${a}",
		})
		ps := NewPropertySources(ConfigTypeLocal, "app")
		_, err := ps.loadFiles(p)
		assert.Error(t, err, `resolve string "\${a}" error << property a not exist`)
	})

	t.Run("loadFiles - Resolve error", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("./testdata/conf/app-${a}.properties")
		_, err := ps.loadFiles(conf.Map(nil))
		assert.Error(t, err, "property a not exist")
	})
}

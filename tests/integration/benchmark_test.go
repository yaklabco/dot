package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yaklabco/dot/tests/integration/testutil"
)

// BenchmarkManage_SinglePackage benchmarks managing a single package.
func BenchmarkManage_SinglePackage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		env := testutil.NewTestEnvironment(b)
		client := testutil.NewTestClient(b, env)

		env.FixtureBuilder().Package("vim").
			WithFile("dot-vimrc", "set nocompatible").
			Create()

		b.StartTimer()
		if err := client.Manage(context.Background(), "vim"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManage_10Packages benchmarks managing 10 packages.
func BenchmarkManage_10Packages(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		env := testutil.NewTestEnvironment(b)
		client := testutil.NewTestClient(b, env)

		// Create 10 packages
		packages := make([]string, 10)
		for j := 0; j < 10; j++ {
			pkgName := filepath.Join("pkg", string(rune('a'+j)))
			packages[j] = pkgName
			env.FixtureBuilder().Package(pkgName).
				WithFile("dot-file", "content").
				Create()
		}

		b.StartTimer()
		if err := client.Manage(context.Background(), packages...); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManage_100Packages benchmarks managing 100 packages.
func BenchmarkManage_100Packages(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		env := testutil.NewTestEnvironment(b)
		client := testutil.NewTestClient(b, env)

		// Create 100 packages
		packages := make([]string, 100)
		for j := 0; j < 100; j++ {
			pkgName := filepath.Join("pkg", string(rune('a'+(j%26))))
			pkgName += string(rune('0' + (j / 26)))
			packages[j] = pkgName
			env.FixtureBuilder().Package(pkgName).
				WithFile("dot-file", "content").
				Create()
		}

		b.StartTimer()
		if err := client.Manage(context.Background(), packages...); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManage_LargeFileTree benchmarks managing package with many files.
func BenchmarkManage_LargeFileTree(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		env := testutil.NewTestEnvironment(b)
		client := testutil.NewTestClient(b, env)

		// Create package with 100 files
		pkg := env.FixtureBuilder().Package("large")
		for j := 0; j < 100; j++ {
			filename := filepath.Join("dot-file", string(rune('a'+(j%26)))+string(rune('0'+(j/26))))
			pkg.WithFile(filename, "content")
		}
		pkg.Create()

		b.StartTimer()
		if err := client.Manage(context.Background(), "large"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRemanage_Unchanged benchmarks remanaging unchanged package.
func BenchmarkRemanage_Unchanged(b *testing.B) {
	b.ReportAllocs()
	// Setup once
	env := testutil.NewTestEnvironment(b)
	client := testutil.NewTestClient(b, env)

	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	if err := client.Manage(context.Background(), "vim"); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// Benchmark remanage
	for i := 0; i < b.N; i++ {
		if err := client.Remanage(context.Background(), "vim"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStatus_Query benchmarks status query.
func BenchmarkStatus_Query(b *testing.B) {
	b.ReportAllocs()
	// Setup once
	env := testutil.NewTestEnvironment(b)
	client := testutil.NewTestClient(b, env)

	// Create and manage 10 packages
	for j := 0; j < 10; j++ {
		pkgName := filepath.Join("pkg", string(rune('a'+j)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	packages := make([]string, 10)
	for j := 0; j < 10; j++ {
		packages[j] = filepath.Join("pkg", string(rune('a'+j)))
	}
	if err := client.Manage(context.Background(), packages...); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// Benchmark status query
	for i := 0; i < b.N; i++ {
		if _, err := client.Status(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkList_Query benchmarks list query.
func BenchmarkList_Query(b *testing.B) {
	b.ReportAllocs()
	// Setup once
	env := testutil.NewTestEnvironment(b)
	client := testutil.NewTestClient(b, env)

	// Create and manage 10 packages
	for j := 0; j < 10; j++ {
		pkgName := filepath.Join("pkg", string(rune('a'+j)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	packages := make([]string, 10)
	for j := 0; j < 10; j++ {
		packages[j] = filepath.Join("pkg", string(rune('a'+j)))
	}
	if err := client.Manage(context.Background(), packages...); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// Benchmark list query
	for i := 0; i < b.N; i++ {
		if _, err := client.List(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnmanage_SinglePackage benchmarks unmanaging a package.
func BenchmarkUnmanage_SinglePackage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		env := testutil.NewTestEnvironment(b)
		client := testutil.NewTestClient(b, env)

		env.FixtureBuilder().Package("vim").
			WithFile("dot-vimrc", "set nocompatible").
			Create()

		if err := client.Manage(context.Background(), "vim"); err != nil {
			b.Fatal(err)
		}

		b.StartTimer()
		if err := client.Unmanage(context.Background(), "vim"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDoctor_HealthCheck benchmarks doctor health check.
func BenchmarkDoctor_HealthCheck(b *testing.B) {
	b.ReportAllocs()
	// Setup once
	env := testutil.NewTestEnvironment(b)
	client := testutil.NewTestClient(b, env)

	// Create and manage packages
	for j := 0; j < 5; j++ {
		pkgName := filepath.Join("pkg", string(rune('a'+j)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	packages := make([]string, 5)
	for j := 0; j < 5; j++ {
		packages[j] = filepath.Join("pkg", string(rune('a'+j)))
	}
	if err := client.Manage(context.Background(), packages...); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// Benchmark doctor
	for i := 0; i < b.N; i++ {
		if _, err := client.Doctor(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}

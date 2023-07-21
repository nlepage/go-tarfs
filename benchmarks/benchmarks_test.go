package tarfs

import (
	"archive/tar"
	"io/fs"
	"math/rand"
	"os"
	"testing"

	"github.com/nlepage/go-tarfs"
)

const chars = "abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVXYZ0123456789"

var (
	randomFileName = make(map[string]string)
)

func TestMain(m *testing.M) {
	rand.Seed(3827653748965)
	generateTarFile("many-small-files.tar", 10000, 1, 10000)
	generateTarFile("few-large-files.tar", 10, 10000000, 100000000)
	os.Exit(m.Run())
}

func BenchmarkOpenTarThenReadFile_ManySmallFiles(b *testing.B) {
	fileName := randomFileName["many-small-files.tar"]

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		openTarThenReadFile("many-small-files.tar", fileName)
	}
}

func BenchmarkOpenTarThenReadFile_FewLargeFiles(b *testing.B) {
	fileName := randomFileName["few-large-files.tar"]

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		openTarThenReadFile("few-large-files.tar", fileName)
	}
}

func BenchmarkReadFile_ManySmallFiles(b *testing.B) {
	benchmarkReadFile(b, "many-small-files.tar")
}

func BenchmarkReadFile_FewLargeFiles(b *testing.B) {
	benchmarkReadFile(b, "few-large-files.tar")
}

func benchmarkReadFile(b *testing.B, tarFileName string) {
	tf, err := os.Open(tarFileName)
	if err != nil {
		panic(err)
	}
	defer tf.Close()

	tfs, err := tarfs.New(tf)
	if err != nil {
		panic(err)
	}

	fileName := randomFileName[tarFileName]

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := fs.ReadFile(tfs, fileName); err != nil {
			panic(err)
		}
	}
}

func openTarThenReadFile(tarName, fileName string) {
	tf, err := os.Open(tarName)
	if err != nil {
		panic(err)
	}
	defer tf.Close()

	var tfs fs.FS

	tfs, err = tarfs.New(tf)
	if err != nil {
		panic(err)
	}

	if _, err := fs.ReadFile(tfs, fileName); err != nil {
		panic(err)
	}
}

func generateTarFile(tarName string, numFiles int, minSize, maxSize int) {
	f, err := os.Create(tarName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := tar.NewWriter(f)
	buf := make([]byte, 1024)
	randomFileIndex := rand.Intn(numFiles)
	defer w.Close()

	for i := 0; i < numFiles; i++ {
		nameLength := rand.Intn(100) + 10
		fileName := ""
		for j := 0; j < nameLength; j++ {
			fileName += string(chars[rand.Intn(len(chars))])
		}

		if i == randomFileIndex {
			randomFileName[tarName] = fileName
		}

		bytesToWrite := rand.Intn(maxSize-minSize) + minSize

		if err := w.WriteHeader(&tar.Header{
			Name:     fileName,
			Typeflag: tar.TypeReg,
			Size:     int64(bytesToWrite),
		}); err != nil {
			panic(err)
		}

		for bytesToWrite != 0 {
			if _, err := rand.Read(buf); err != nil {
				panic(err)
			}

			if bytesToWrite < 1024 {
				if _, err := w.Write(buf[:bytesToWrite]); err != nil {
					panic(err)
				}
				bytesToWrite = 0
			} else {
				if _, err := w.Write(buf); err != nil {
					panic(err)
				}
				bytesToWrite -= 1024
			}
		}
	}
}

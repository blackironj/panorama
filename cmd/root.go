package cmd

import (
	"fmt"
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uilive"
	"github.com/spf13/cobra"

	"github.com/blackironj/panorama/conv"
)

const (
	defaultEdgeLen     = 1024
	maxConcurrentFiles = 10 // Adjust this number based on your system's file descriptor limit
	defaultJpegQuality = 75
)

var (
	inFilePath string
	outFileDir string
	inDirPath  string
	edgeLen    int
	sides      []string
	quality    int

	validSides = []string{"front", "back", "left", "right", "top", "bottom"}
	semaphore  = make(chan struct{}, maxConcurrentFiles)
	progress   = struct {
		sync.Mutex
		totalFiles     int
		processedFiles int
		startTime      time.Time
		errors         []string
	}{}
	rootCmd = &cobra.Command{
		Use:   "panorama",
		Short: "convert equirectangular panorama img to Cubemap img",
		Run: func(cmd *cobra.Command, args []string) {
			if inFilePath == "" && inDirPath == "" {
				er("Need an input image file path or input directory")
			}
			if len(inFilePath) > 0 && len(inDirPath) > 0 {
				er("Need only one path, not both")
			}

			progress.startTime = time.Now()
			fmt.Println("Start conversion.")
			if inFilePath != "" {
				progress.totalFiles = 1
				processSingleImage(inFilePath, outFileDir, false)
			} else {
				processDirectory(inDirPath, outFileDir)
			}
			elapsed := time.Since(progress.startTime).Seconds()
			fmt.Printf("Processing complete. elapsed: %.2f sec\n\n", elapsed)

			if len(progress.errors) > 0 {
				fmt.Println("\nErrors:")
				for _, err := range progress.errors {
					fmt.Println(err)
				}
			}
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&inFilePath, "in", "i", "", "input image file path (required if --indir is not specified)")
	rootCmd.Flags().StringVarP(&inDirPath, "indir", "d", "", "input directory path (required if --in is not specified)")
	rootCmd.Flags().StringVarP(&outFileDir, "out", "o", ".", "output file directory path")
	rootCmd.Flags().IntVarP(&edgeLen, "len", "l", defaultEdgeLen, "edge length of a cube face")
	rootCmd.Flags().StringSliceVarP(&sides, "sides", "s", []string{}, "array of sides [front,back,left,right,top,bottom] (default: all sides)")
	rootCmd.Flags().IntVarP(&quality, "quality", "q", defaultJpegQuality, "jpeg file output quality ranges from 1 to 100 inclusive, higher is better")
}

func processSingleImage(inPath, outDir string, needSubdir bool) {
	semaphore <- struct{}{}        // Acquire a semaphore
	defer func() { <-semaphore }() // Release the semaphore when done

	inImage, ext, err := conv.ReadImage(inPath)
	if err != nil {
		progress.Lock()
		progress.errors = append(progress.errors, fmt.Sprintf("Error reading image %s: %v", inPath, err))
		progress.Unlock()
		return
	}

	if len(sides) == 0 {
		sides = validSides
	} else {
		for _, side := range sides {
			if !isValidSide(side) {
				er(fmt.Sprintf("Invalid side specified: %s. Valid sides are %v", side, validSides))
			}
		}
	}

	canvases, err := safeConvertEquirectangularToCubeMap(edgeLen, inImage, sides)

	if err != nil {
		progress.Lock()
		progress.errors = append(progress.errors, fmt.Sprintf("Error converting image %s: %v", inPath, err))
		progress.Unlock()
		return
	}

	if needSubdir {
		outDir = filepath.Join(outDir, strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath)))
	}
	if err := conv.WriteImage(canvases, outDir, ext, sides, quality); err != nil {
		progress.Lock()
		progress.errors = append(progress.errors, fmt.Sprintf("Error writing images for %s: %v", inPath, err))
		progress.Unlock()
		return
	}

	progress.Lock()
	progress.processedFiles++
	progress.Unlock()
}

func processDirectory(inDir, outDir string) {
	files, err := os.ReadDir(inDir)
	if err != nil {
		er(err)
	}

	progress.totalFiles = len(files)

	writer := uilive.New()
	writer.Start()
	defer writer.Stop()

	var wg sync.WaitGroup
	for _, file := range files {
		if !file.IsDir() && isImageFile(file) {
			wg.Add(1)
			go func(file fs.DirEntry) {
				defer wg.Done()
				inPath := filepath.Join(inDir, file.Name())
				processSingleImage(inPath, outDir, true)
				updateProgress(writer)
			}(file)
		}
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			updateProgress(writer)
			progress.Lock()
			remaining := progress.totalFiles - progress.processedFiles
			if remaining <= 0 {
				progress.Unlock()
				break
			}
			progress.Unlock()
		}
	}()

	wg.Wait()
}

func updateProgress(writer *uilive.Writer) {
	progress.Lock()
	defer progress.Unlock()
	remaining := progress.totalFiles - progress.processedFiles
	elapsed := time.Since(progress.startTime).Seconds()
	eta := float64(remaining) / (float64(progress.processedFiles) / elapsed)
	fmt.Fprintf(writer, "Progress: %d/%d files processed. ETA: %.2f seconds. IT/S: %.2f\n", progress.processedFiles, progress.totalFiles, eta, float64(progress.processedFiles)/elapsed)
}

func isImageFile(file fs.DirEntry) bool {
	ext := strings.ToLower(filepath.Ext(file.Name()))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		er(err)
	}
}

func isValidSide(side string) bool {
	for _, s := range validSides {
		if s == side {
			return true
		}
	}
	return false
}

func safeConvertEquirectangularToCubeMap(edgeLen int, imgIn image.Image, sides []string) ([]*image.RGBA, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in safeConvertEquirectangularToCubeMap: %v\n", r)
		}
	}()
	return conv.ConvertEquirectangularToCubeMap(edgeLen, imgIn, sides), nil
}

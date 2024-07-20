package cmd

import (
	"fmt"
	"os"
	"time"
	"image"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	"github.com/blackironj/panorama/conv"
)

const defaultEdgeLen = 1024

var (
	inFilePath string
	outFileDir string
	edgeLen    int
	sides      []string

	validSides = []string{"front", "back", "left", "right", "top", "bottom"}

	rootCmd = &cobra.Command{
		Use:   "panorama",
		Short: "convert equirectangular panorama img to Cubemap img",
		Run: func(cmd *cobra.Command, args []string) {
			if inFilePath == "" {
				er("Need an image for converting")
			}

			fmt.Println("Reading the image...")
			inImage, ext, err := conv.ReadImage(inFilePath)
			if err != nil {
				er(err)
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

			s := spinner.New(spinner.CharSets[33], 100*time.Millisecond)
			s.FinalMSG = "Complete converting!\n"
			s.Prefix = "Converting..."

			s.Start()
			canvases, err := safeConvertEquirectangularToCubeMap(edgeLen, inImage, sides)
			s.Stop()
			if err != nil {
				er(err)
			}

			fmt.Println("Writing images...")
			if err := conv.WriteImage(canvases, outFileDir, ext, sides); err != nil {
				er(err)
			}
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&inFilePath, "in", "i", "", "input image file path (required)")
	rootCmd.Flags().StringVarP(&outFileDir, "out", "o", ".", "output file directory path")
	rootCmd.Flags().IntVarP(&edgeLen, "len", "l", defaultEdgeLen, "edge length of a cube face")
	rootCmd.Flags().StringSliceVarP(&sides, "sides", "s", []string{}, "array of sides [front,back,left,right,top,bottom] (default: all sides)")
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

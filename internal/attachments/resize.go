package attachments

import "image"

func resizeNearest(src image.Image, dstW, dstH int) image.Image {
	srcBounds := src.Bounds()
	srcW, srcH := srcBounds.Dx(), srcBounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		srcY := srcBounds.Min.Y + y*srcH/dstH
		for x := 0; x < dstW; x++ {
			srcX := srcBounds.Min.X + x*srcW/dstW
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

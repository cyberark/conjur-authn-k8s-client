package utils

import (
	"strings"
)

func ByteSlicePrintf(tmplt string, separator string, subByteSlices ...[]byte) []byte {
	tmpltChunks := strings.Split(tmplt, separator)
	tmpltChunkByteSlices := make([][]byte, len(tmpltChunks))
	resultingByteSliceSize := 0

	for i, tmpltChunk := range tmpltChunks {
		chunkByte := []byte(tmpltChunk)
		tmpltChunkByteSlices[i] = chunkByte
		resultingByteSliceSize += len(chunkByte)
		if i < len(subByteSlices) {
			resultingByteSliceSize += len(subByteSlices[i])
		}
	}

	resultingByteSlice := make([]byte, resultingByteSliceSize)

	resultingByteSliceIndex := 0
	for tmpltChunkByteSliceIndex, tmpltChunkByteSlice := range tmpltChunkByteSlices {
		for _, tmpltChunkByteSliceByte := range tmpltChunkByteSlice {
			resultingByteSlice[resultingByteSliceIndex] = tmpltChunkByteSliceByte
			resultingByteSliceIndex += 1
		}

		if tmpltChunkByteSliceIndex < len(subByteSlices) {
			subByteSlice := subByteSlices[tmpltChunkByteSliceIndex]
			for _, subByteSliceByte := range subByteSlice {
				resultingByteSlice[resultingByteSliceIndex] = subByteSliceByte
				resultingByteSliceIndex += 1
			}
		}
	}

	return resultingByteSlice
}

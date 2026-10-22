package document

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

type DocumentChunker struct {
	ChunkSize int
	Overlap   int
}

func NewDocumentChunker(chunkSize, overlap int) *DocumentChunker {
	return &DocumentChunker{
		ChunkSize: chunkSize,
		Overlap:   overlap,
	}
}

// ParseUploadedFile extracts raw text from HTTP multipart uploaded files (.txt or .pdf)
func (c *DocumentChunker) ParseUploadedFile(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	switch ext {
	case ".txt":
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, file); err != nil {
			return "", err
		}
		return buf.String(), nil

	case ".pdf":
		// Read entire PDF buffer into memory for parsing
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, file); err != nil {
			return "", err
		}
		readerAt := bytes.NewReader(buf.Bytes())
		
		pdfReader, err := pdf.NewReader(readerAt, int64(buf.Len()))
		if err != nil {
			return "", fmt.Errorf("failed to parse PDF: %w", err)
		}

		var sb strings.Builder
		totalPages := pdfReader.NumPage()
		for i := 1; i <= totalPages; i++ {
			p := pdfReader.Page(i)
			if p.V.IsNull() {
				continue
			}
			text, err := p.GetPlainText(nil)
			if err != nil {
				continue
			}
			sb.WriteString(text)
			sb.WriteString("\n")
		}
		return sb.String(), nil

	default:
		return "", errors.New("unsupported file format. Only .txt and .pdf are allowed")
	}
}

// ChunkText splits raw text into overlapping token/word chunks
func (c *DocumentChunker) ChunkText(text string) []string {
	words := strings.Fields(text)
	var chunks []string

	if len(words) == 0 {
		return chunks
	}

	step := c.ChunkSize - c.Overlap
	if step <= 0 {
		step = c.ChunkSize
	}

	for i := 0; i < len(words); i += step {
		end := i + c.ChunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
		if end == len(words) {
			break
		}
	}

	return chunks
}
package vectorstore

import (
	"context"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VectorStore struct {
	client     *qdrant.Client
	collection string
}

type DocumentMetadata struct {
	DocumentID   string `json:"document_id"`
	DocumentName string `json:"document_name"`
	DocumentSize int64  `json:"document_size"`
	ChunksCount  int    `json:"chunks_count"`
}

func NewVectorStore(host string, port int, collection string) (*VectorStore, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
		GrpcOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	})
	if err != nil {
		return nil, err
	}

	return &VectorStore{
		client:     client,
		collection: collection,
	}, nil
}

func (vs *VectorStore) EnsureCollection(ctx context.Context, vectorSize uint64) error {
	exists, err := vs.client.CollectionExists(ctx, vs.collection)
	if err != nil {
		return err
	}

	if !exists {
		return vs.client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: vs.collection,
			VectorsConfig: &qdrant.VectorsConfig{
				Config: &qdrant.VectorsConfig_Params{
					Params: &qdrant.VectorParams{
						Size:     vectorSize,
						Distance: qdrant.Distance_Cosine,
					},
				},
			},
		})
	}
	return nil
}

func (vs *VectorStore) UpsertChunks(ctx context.Context, chunks []string, embeddings [][]float32, documentID string, documentName string, documentSize int64) error {
	var points []*qdrant.PointStruct

	for i, chunk := range chunks {
		pointID := uuid.New().String()
		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(pointID),
			Vectors: qdrant.NewVectors(embeddings[i]...),
			Payload: qdrant.NewValueMap(map[string]any{
				"text":          chunk,
				"document_id":   documentID,
				"document_name": documentName,
				"document_size": documentSize,
			}),
		})
	}

	_, err := vs.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: vs.collection,
		Points:         points,
	})
	return err
}

func (vs *VectorStore) Search(ctx context.Context, queryVector []float32, topK uint64, documentIDs []string) ([]string, error) {
	var filter *qdrant.Filter
	if len(documentIDs) > 0 {
		conditions := make([]*qdrant.Condition, len(documentIDs))
		for i, id := range documentIDs {
			conditions[i] = qdrant.NewMatch("document_id", id)
		}
		filter = &qdrant.Filter{
			Should: conditions,
		}
	}

	searchResult, err := vs.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: vs.collection,
		Query:          qdrant.NewQueryDense(queryVector),
		Filter:         filter,
		Limit:          &topK,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	var contextChunks []string
	for _, point := range searchResult {
		if textVal, ok := point.Payload["text"]; ok {
			contextChunks = append(contextChunks, textVal.GetStringValue())
		}
	}

	return contextChunks, nil
}

// ListDocuments scrolls through points to aggregate unique document IDs and chunk counts
func (vs *VectorStore) ListDocuments(ctx context.Context) ([]DocumentMetadata, error) {
	scrollResult, err := vs.client.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: vs.collection,
		WithPayload:    qdrant.NewWithPayload(true),
		Limit:          qdrant.PtrOf(uint32(10000)),
	})
	if err != nil {
		return nil, err
	}

	docCounts := make(map[string]int)
	docNames := make(map[string]string)
	docSizes := make(map[string]int64)

	for _, point := range scrollResult {
		if val, ok := point.Payload["document_id"]; ok {
			docID := val.GetStringValue()
			if docID != "" {
				docCounts[docID]++
				if nameVal, ok := point.Payload["document_name"]; ok {
					docNames[docID] = nameVal.GetStringValue()
				}
				if sizeVal, ok := point.Payload["document_size"]; ok {
					docSizes[docID] = sizeVal.GetIntegerValue()
				}
			}
		}
	}

	var documents []DocumentMetadata
	for docID, count := range docCounts {
		name := docNames[docID]
		if name == "" {
			name = "Unknown"
		}
		documents = append(documents, DocumentMetadata{
			DocumentID:   docID,
			DocumentName: name,
			DocumentSize: docSizes[docID],
			ChunksCount:  count,
		})
	}

	return documents, nil
}

// DeleteByDocumentID deletes all vector points matching the given document ID filter
func (vs *VectorStore) DeleteByDocumentID(ctx context.Context, documentID string) error {
	_, err := vs.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: vs.collection,
		Points: qdrant.NewPointsSelectorFilter(
			&qdrant.Filter{
				Must: []*qdrant.Condition{
					qdrant.NewMatch("document_id", documentID),
				},
			},
		),
	})
	return err
}

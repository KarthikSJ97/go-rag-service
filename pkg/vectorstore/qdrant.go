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

func (vs *VectorStore) UpsertChunks(ctx context.Context, chunks []string, embeddings [][]float32, documentID string) error {
	var points []*qdrant.PointStruct

	for i, chunk := range chunks {
		pointID := uuid.New().String()
		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(pointID),
			Vectors: qdrant.NewVectors(embeddings[i]...),
			Payload: qdrant.NewValueMap(map[string]any{
				"text":        chunk,
				"document_id": documentID,
			}),
		})
	}

	_, err := vs.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: vs.collection,
		Points:         points,
	})
	return err
}

func (vs *VectorStore) Search(ctx context.Context, queryVector []float32, topK uint64) ([]string, error) {
	searchResult, err := vs.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: vs.collection,
		Query:          qdrant.NewQuery(queryVector...),
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
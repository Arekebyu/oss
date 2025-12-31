package api

import (
	"context"
	"fmt"
	"time"

	"oss/internal/models"
	"oss/internal/search"
	pb "oss/pb"
)

type SearchService struct {
	ESClient *search.Client
	MLClient pb.MLServiceClient
}

type Result struct {
	Title string  `json:"title"`
	URL   string  `json:"url"`
	Score float64 `json:"score"`
	Text  string  `json:"text"`
}

func (s *SearchService) SearchAndRank(ctx context.Context, query string) ([]Result, error) {
	candidates, err := s.ESClient.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("elastic search failed: %w", err)
	}
	if len(candidates) == 0 {
		return []Result{}, nil
	}

	var pbCandidates []*pb.Document
	docMap := make(map[string]models.ScrapedPage)

	for _, doc := range candidates {
		docMap[doc.URL] = doc

		snippet := doc.Title
		if len(doc.Sections) > 0 {
			snippet += " " + doc.Sections[0].Content
		}

		pbCandidates = append(pbCandidates, &pb.Document{
			Id:             doc.URL,
			Title:          doc.Title,
			ContentSnippet: snippet,
		})
	}

	rankCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rankReq := &pb.RankRequest{
		Query:      query,
		Candidates: pbCandidates,
	}

	rankRes, err := s.MLClient.ReRank(rankCtx, rankReq)
	if err != nil {
		// fallback to elasticsearch ranking if python service is down
		fmt.Printf(" ML Service failed: %v. Returning keyword results.\n", err)
		return s.mapToResults(candidates, nil), nil
	}

	return s.mapToResults(nil, rankRes.Results, docMap), nil
}

func (s *SearchService) mapToResults(
	original []models.ScrapedPage,
	ranked []*pb.RankedDocument,
	lookup ...map[string]models.ScrapedPage,
) []Result {
	var final []Result

	// fallback mode
	if ranked == nil {
		for _, page := range original {
			final = append(final, Result{
				Title: page.Title,
				URL:   page.URL,
				Score: 1.0,
				Text:  page.Sections[0].Content,
			})
		}
		return final
	}

	docMap := lookup[0]
	for _, hit := range ranked {
		originalDoc := docMap[hit.Id]
		final = append(final, Result{
			Title: originalDoc.Title,
			URL:   originalDoc.URL,
			Score: float64(hit.Score),
			Text:  originalDoc.Sections[0].Content,
		})
	}
	return final
}

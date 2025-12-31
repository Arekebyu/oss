import axios from 'axios';

const API_URL = "http://localhost:8080";

export interface SearchResult {
  title: string;
  url: string;
  score: number;
  text: string;
}

export interface SearchResponse {
  query: string;
  count: number;
  results: SearchResult[];
}

export const searchDocs = async (query: string): Promise<SearchResponse> => {
  const response = await axios.get(`${API_URL}/search`, {
    params: { q: query },
  });
  return response.data;
};
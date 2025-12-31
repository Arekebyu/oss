'use client'

import { useEffect, useState } from "react";
import { searchDocs, SearchResult } from "@/lib/api";
import { loadEnvConfig } from "@next/env";

export default function Home() {
  const projDir = process.cwd()

  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const [searchedQuery, setSearchedQuery] = useState('')

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setLoading(true)
    setSearched(true)

    try {
      const data = await searchDocs(query);
      setResults(data.results || []);
    } catch (err) {
      console.error("search failed", err)
    } finally {
      setLoading(false)
    }
  };

  useEffect(() => {
    setSearchedQuery(query)
  }, results)

  return (
    <main className="min-h-screen bg-gray-50 flex flex-col items-center py-10">
      <div className={`w-full max-w-3xl px-4 transition-all duration-500 ${searched ? 'mt-0' : 'mt-[20vh]'}`}>
        <h1 className={`text-4xl font-bold text-gray-800 text-center mb-8 ${searched ? 'hidden' : 'block'}`}>
          Open Source Search
        </h1>

        <form onSubmit={handleSearch} className="w-full relative">
          <input
            type="text"
            className="w-full p-4 pl-6 rounded-full border border-gray-300 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 text-lg text-gray-500"
            placeholder="Search documentation (e.g. 'Reshape tensor pytorch')..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <button 
            type="submit"
            className="absolute right-3 top-3 bg-blue-600 text-white px-6 py-2 rounded-full hover:bg-blue-700 transition"
            disabled={loading}
          >
            {loading ? '...' : 'Search'}
          </button>
        </form>
      </div>

      <div className="w-full max-w-3xl px-4 mt-10 space-y-6">
        {results.map((result, index) => (
          <div key={index} className="bg-white p-6 rounded-lg shadow-sm border border-gray-100 hover:shadow-md transition">
            
            <a href={result.url} target="_blank" rel="noopener noreferrer" className="group">
              <h2 className="text-xl font-semibold text-blue-700 group-hover:underline">
                {result.title}
              </h2>
              <p className="text-sm text-green-700 tfalseruncate mb-2">{result.url}</p>
            </a>

            <p className="text-gray-600 text-sm leading-relaxed">
              {result.text}
            </p>
            
            <div className="mt-3 flex items-center gap-2">
              <span className={`text-xs px-2 py-1 rounded font-mono font-medium ${
                result.score > 0.8 ? 'bg-green-100 text-green-800' : 
                result.score > 0.5 ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-800'
              }`}>
                Relevance: {(result.score * 100).toFixed(1)}%
              </span>
              <span className="text-xs text-gray-400">
                (Ranked by Cross-Encoder)
              </span>
            </div>
          </div>
        ))}

        {searched && !loading && results.length === 0 && (
          <p className="text-center text-gray-500">No results found for "{query}"</p>
        )}
      </div>

    </main>
  );
}

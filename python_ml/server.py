import grpc
import os
from concurrent import futures
import time
import logging

from sentence_transformers import CrossEncoder

import search_ml_pb2
import search_ml_pb2_grpc

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class MLServiceHandler(search_ml_pb2_grpc.MLServiceServicer):
    def __init__(self):
        logger.info("loading model")
        self.model = CrossEncoder('cross-encoder/ms-marco-MiniLM-L6-v2')
        logger.info("model loaded")

    def ReRank(self, request, context):
        start_time = time.time()
        logger.info(f"Ranking {len(request.candidates)} documents for query: \"{request.query}\"")
        if not request.candidates:
            return search_ml_pb2.RankResponse(results=[])        

        X = []

        for doc in request.candidates:
            # bert unfortuantely has a small token cap of 512, so, we have to truncate input
            X.append([request.query, f"{doc.title} {doc.content_snippet}"[:1000]])
        
        scores = self.model.predict(X)
        
        ranked_results = []

        for i, score in enumerate(scores):
            ranked_results.append(
                search_ml_pb2.RankedDocument(
                    id=request.candidates[i].id,
                    score=float(score)
                )
            )

        ranked_results.sort(key=lambda x: x.score, reverse=True)

        elapsed = time.time() - start_time
        logger.info(f"ranking completed in {elapsed:.4f}s")

        return search_ml_pb2.RankResponse(results=ranked_results)
    
    def GetEmbedding(self, request, context):
        return search_ml_pb2.EmbeddingResponse(vector=[0.1, 0.2, 0.3])

PORT = os.environ.get("port", "50051")

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    search_ml_pb2_grpc.add_MLServiceServicer_to_server(MLServiceHandler(), server)
    server.add_insecure_port(f'[::]:{PORT}')
    print(f'ML server started on {PORT}')
    server.start()
    
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()

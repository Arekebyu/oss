import grpc
from concurrent import futures
import time

import search_ml_pb2
import search_ml_pb2_grpc

class MLServiceHandler(search_ml_pb2_grpc.MLServiceServicer):
    def ReRank(self, request, context):
        print("received context" + request.query)
        
        ranked_results = []
        query_words = request.query.lower().split()

        for doc in request.candidates:
            # Todo!(), simple matching for now
            score = 0.0
            content = doc.content_snippet.lower()

            matches = sum(1 for word in query_words if word in content)
            if len(query_words) > 0 :
                score = matches / len(query_words)
            # logic end
            ranked_results.append(
                search_ml_pb2.RankedDocument(id=doc.id, score=score)
            )
        
        ranked_results.sort(key=lambda x: x.score, reverse=True)

        return search_ml_pb2.RankResponse(results=ranked_results)
    
    def GetEmbedding(self, request, context):
        return search_ml_pb2.EmbeddingResponse(vector=[0.1, 0.2, 0.3])

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    search_ml_pb2_grpc.add_MLServiceServicer_to_server(MLServiceHandler(), server)
    server.add_insecure_port('[::]:50051')
    print("ML server started on 50051")
    server.start()
    
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)
if __name__ == '__main__':
    serve()

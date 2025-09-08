import asyncio
from typing import Callable, Optional

class PubSubClient:
    def __init__(self, socket_path: str = "/tmp/pubsub.sock"):
        self.socket_path = socket_path
        self.reader: Optional[asyncio.StreamReader] = None
        self.writer: Optional[asyncio.StreamWriter] = None
        self.subscriptions = {}
        
    async def connect(self):
        print(f"Connecting to {self.socket_path}...")
        
        
    async def publish(self, topic: str, message: str):
        pass
        
    async def subscribe(self, topic: str, callback: Callable):
        pass

    #Implement connection publishing and subscription
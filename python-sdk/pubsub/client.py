import asyncio
import struct
from typing import Callable, Dict, Optional
import logging
from .bloom import TopicFilter, create_publish_message

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class PubSubClient:    
    def __init__(self, socket_path: str = "/tmp/pubsub.sock"):
        self.socket_path = socket_path
        self.reader: Optional[asyncio.StreamReader] = None
        self.writer: Optional[asyncio.StreamWriter] = None
        self.subscriptions: Dict[str, Callable] = {}
        self.topic_filter = TopicFilter()
        self.connected = False
        self._receive_task = None
        
    async def connect(self):
        try:
            self.reader, self.writer = await asyncio.open_unix_connection(
                self.socket_path
            )
            self.connected = True
            logger.info(f"Connected to daemon at {self.socket_path}")
            
            self._receive_task = asyncio.create_task(self._receive_loop())
            
        except Exception as e:
            logger.error(f"Failed to connect: {e}")
            raise
    
    async def disconnect(self):
        self.connected = False
        
        if self._receive_task:
            self._receive_task.cancel()
            try:
                await self._receive_task
            except asyncio.CancelledError:
                pass
        
        if self.writer:
            self.writer.close()
            await self.writer.wait_closed()
            
        logger.info("Disconnected from daemon")
    
    async def publish(self, topic: str, message: str):
        if not self.connected:
            raise RuntimeError("Not connected to daemon")
        
        payload = f"{topic}|{message}"
        
        if len(payload) > 61:
            max_msg_len = 61 - len(topic) - 1
            message = message[:max_msg_len]
            payload = f"{topic}|{message}"
        
        cmd = f"PUB {payload}"
        self.writer.write(cmd.encode())
        await self.writer.drain()
        
        logger.debug(f"Published: {topic} -> {message}")
    
    async def subscribe(self, topic: str, callback: Callable):
        if not self.connected:
            raise RuntimeError("Not connected to daemon")
        
        self.subscriptions[topic] = callback
        self.topic_filter.add_topic(topic)
        
        await self._send_filter()
        
        logger.info(f"Subscribed to: {topic}")
    
    async def unsubscribe(self, topic: str):
        if topic in self.subscriptions:
            del self.subscriptions[topic]
            self.topic_filter.remove_topic(topic)
            await self._send_filter()
            logger.info(f"Unsubscribed from: {topic}")
    
    async def _send_filter(self):
        filter_bytes = self.topic_filter.serialize()
        cmd = b"SUB " + filter_bytes
        
        self.writer.write(cmd)
        await self.writer.drain()
    
    async def _receive_loop(self):
        buffer = b""
        
        while self.connected:
            try:
                data = await self.reader.read(1024)
                if not data:
                    break
                
                buffer += data
                
                while b"\n" in buffer or len(buffer) > 100:
                    if b"\n" in buffer:
                        line, buffer = buffer.split(b"\n", 1)
                    else:
                        line = buffer
                        buffer = b""
                    
                    await self._process_message(line)
                    
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Receive error: {e}")
                break
    
    async def _process_message(self, data: bytes):
        try:
            message = data.decode('utf-8')
            
            if message.startswith("MSG "):
                payload = message[4:]
                parts = payload.split("|", 1)
                
                if len(parts) == 2:
                    topic, msg = parts
                    
                    if topic in self.subscriptions:
                        callback = self.subscriptions[topic]
                        if asyncio.iscoroutinefunction(callback):
                            await callback(msg)
                        else:
                            callback(msg)
                            
        except Exception as e:
            logger.error(f"Error processing message: {e}")
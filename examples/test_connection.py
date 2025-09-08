import asyncio
import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from pubsub import PubSubClient

async def main():
    client = PubSubClient()
    print("Client created successfully!")
    # await client.connect()  # Will implement this next

if __name__ == "__main__":
    asyncio.run(main())
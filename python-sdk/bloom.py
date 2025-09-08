import struct
from typing import List, Set
from pybloom_live import BloomFilter


class TopicFilter:
    FILTER_SIZE = 1000
    ERROR_RATE = 0.01
    
    def __init__(self):
        self.topics: Set[str] = set()
        self.bloom: BloomFilter = BloomFilter(
            capacity=self.FILTER_SIZE,
            error_rate=self.ERROR_RATE
        )
    
    def add_topic(self, topic: str) -> None:
        if topic not in self.topics:
            self.topics.add(topic)
            self.bloom.add(topic)
    
    def remove_topic(self, topic: str) -> None:
        if topic in self.topics:
            self.topics.discard(topic)
            self._regenerate_filter()
    
    def _regenerate_filter(self) -> None:
        self.bloom = BloomFilter(
            capacity=self.FILTER_SIZE,
            error_rate=self.ERROR_RATE
        )
        for topic in self.topics:
            self.bloom.add(topic)
    
    def test(self, topic: str) -> bool:
        return topic in self.bloom
    
    def serialize(self) -> bytes:
        bitarray = self.bloom.bitarray

        metadata = struct.pack(
            len(bitarray),
            self.bloom.hash_k
        )
        
        bloom_bytes = bitarray.tobytes()
        
        return metadata + bloom_bytes
    
    def deserialize(cls, data: bytes) -> 'TopicFilter':

        bit_size, hash_k = struct.unpack('!II', data[:8])
        
        bloom_bytes = data[8:]
        filter_obj = cls()
        return filter_obj
    
    def clear(self) -> None:
        self.topics.clear()
        self.bloom = BloomFilter(
            capacity=self.FILTER_SIZE,
            error_rate=self.ERROR_RATE
        )
    
    def get_topics(self) -> List[str]:
        return list(self.topics)
    
    def __len__(self) -> int:
        return len(self.topics)
    
    def __contains__(self, topic: str) -> bool:
        return self.test(topic)
    
    def __repr__(self) -> str:
        return f"TopicFilter(topics={len(self.topics)}, size={self.FILTER_SIZE}, error={self.ERROR_RATE})"



def create_subscription_message(topics: List[str]) -> bytes:
    filter_obj = TopicFilter()
    for topic in topics:
        filter_obj.add_topic(topic)
    
    filter_bytes = filter_obj.serialize()
    message = b"SUB " + filter_bytes
    
    return message


def create_publish_message(topic: str, message: str) -> bytes:
    MAX_PAYLOAD_SIZE = 61
    
    payload = f"{topic}|{message}"
    
    if len(payload) > MAX_PAYLOAD_SIZE:
        topic_and_sep_len = len(topic) + 1
        if topic_and_sep_len >= MAX_PAYLOAD_SIZE:
            raise ValueError(f"Topic too long: {len(topic)} bytes (max: {MAX_PAYLOAD_SIZE - 2})")
        
        max_msg_len = MAX_PAYLOAD_SIZE - topic_and_sep_len
        message = message[:max_msg_len]
        payload = f"{topic}|{message}"
    
    return f"PUB {payload}".encode('utf-8')


if __name__ == "__main__":
    filter_obj = TopicFilter()
    
    # Add some topics
    test_topics = [
        "sensors/temperature",
        "sensors/humidity", 
        "devices/light",
        "devices/door"
    ]
    
    for topic in test_topics:
        filter_obj.add_topic(topic)
    
    print(f"Filter created: {filter_obj}")
    print(f"Topics: {filter_obj.get_topics()}")
    
    print("\nTesting membership:")
    print(f"'sensors/temperature' in filter: {'sensors/temperature' in filter_obj}")
    print(f"'sensors/pressure' in filter: {'sensors/pressure' in filter_obj}")
    
    serialized = filter_obj.serialize()
    print(f"\nSerialized size: {len(serialized)} bytes")
    
    pub_msg = create_publish_message("test/topic", "Hello World!")
    print(f"\nPublish message: {pub_msg}")
    print(f"Message size: {len(pub_msg)} bytes")
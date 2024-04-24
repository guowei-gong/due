package buffer

var defaultWriterPool = NewWriterPool([]int{32, 64, 128, 256, 512, 1024, 2048, 4096, 10240})

type NocopyBuffer struct {
	len  int
	head *NocopyNode
	tail *NocopyNode
}

var _ Buffer = &NocopyBuffer{}

func NewNocopyBuffer() *NocopyBuffer {
	return &NocopyBuffer{len: -1}
}

// Len 获取字节长度
func (b *NocopyBuffer) Len() int {
	if b.len >= 0 {
		return b.len
	}

	n := 0
	for node := b.head; node != nil; {
		n += node.Len()
		node = node.next
	}
	b.len = n

	return n
}

// Mount 挂载数据到Buffer上
func (b *NocopyBuffer) Mount(data []byte, whence ...Whence) {
	if len(whence) > 0 && whence[0] == Head {
		b.addToHead(&NocopyNode{buf: data})
	} else {
		b.addToTail(&NocopyNode{buf: data})
	}
}

// Malloc 分配一块内存给Writer
func (b *NocopyBuffer) Malloc(cap int, whence ...Whence) *Writer {
	writer := defaultWriterPool.Get(cap)

	if len(whence) > 0 && whence[0] == Head {
		b.addToHead(&NocopyNode{buf: writer, pool: defaultWriterPool})
	} else {
		b.addToTail(&NocopyNode{buf: writer, pool: defaultWriterPool})
	}

	return writer
}

// Range 迭代
func (b *NocopyBuffer) Range(fn func(node *NocopyNode) bool) {
	node := b.head
	for node != nil {
		next := node.next

		if fn(node) {
			node = next
		} else {
			break
		}
	}
}

// Release 释放
func (b *NocopyBuffer) Release() {
	node := b.head
	for node != nil {
		next := node.next
		node.Release()
		node = next
	}
	b.head = nil
	b.tail = nil
}

// 添加到尾部
func (b *NocopyBuffer) addToTail(node *NocopyNode) {
	if b.head == nil {
		b.head = node
		b.tail = node
	} else {
		b.tail.next = node
		b.tail.next.prev = b.tail
		b.tail = node
	}

	b.len = -1
}

// 添加到头部
func (b *NocopyBuffer) addToHead(node *NocopyNode) {
	if b.head == nil {
		b.head = node
		b.tail = node
	} else {
		node.next = b.head
		b.head.prev = node
		b.head = node
	}

	b.len = -1
}
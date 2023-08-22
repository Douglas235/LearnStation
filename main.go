package main

import "LearnStation/BloomFilter"

// 初始化过滤器时，需要知道对应的业务场景有多少元素（期望的容量），
// 以及期望的误判概率。常见的误判率为 1%, 误报率越低，需要的内存就越多，
// 同时容量越大，需要的内存就越多
func main() {
    filter := BloomFilter.NewWithEstimates(1000000, 0.01)
    hw := []byte(`hello world`)
    hg := []byte(`hello golang`)
    filter.Add(hw)
    println(filter.Test(hw))
    println(filter.Test(hg))
}

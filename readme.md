# redisrate

    使用redi-go连接池实现的分布式频率限制组件
    该频率限制组件，需要把redis pool 传递给NewLimiter函数创建一个redis limiter.


# 参考组件 

    参考https://github.com/go-redis/redis_rate 实现的分布式限速组件

# 关于redigo pool 封装

    可以参考 https://github.com/daheige/thinkgo/blob/master/rediscache/redis.go#L47

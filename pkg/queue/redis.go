package queue

import (
    "context"
    "encoding/json"
    "github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

const StreamName = "image_jobs"
const GroupName = "image_workers"

func InitRedis(addr string) error {
    Rdb = redis.NewClient(&redis.Options{
        Addr: addr,
    })

    // Create stream group if not exists
    err := Rdb.XGroupCreateMkStream(Ctx, StreamName, GroupName, "$").Err()
    if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
        return err
    }
    return nil
}

func PublishImageJob(job ImageJob) error {
    body, _ := json.Marshal(job)

    _, err := Rdb.XAdd(Ctx, &redis.XAddArgs{
        Stream: StreamName,
        Values: map[string]interface{}{
            "job": string(body),
        },
    }).Result()
    return err
}

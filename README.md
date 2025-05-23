# Bluesky Lonely Posts

'Lonely Posts' was a Bluesky feed that displayed posts with zero interactions (i.e. no replies, likes, or quotes).

In this repo, there are to Go services:

1. Intake: responsible for collecting posts from the Bluesky Jetstream and storing them in Valkey
2. Server: responsible for serving the feed, and reading posts from Valkey

The logic was pretty staightforward:

- Add posts to Valkey as they appear in the Jetstream
- If we encounter a reply/quote/like for a post, delete it the post from Valkey

This repo also contains the infra to run these services on ECS Fargate. I shut down the feed because it didn't prove to be valuable, and it cost $15/month to run.

Podman notes:

```
podman machine set --cpus 2 --memory 4096
```

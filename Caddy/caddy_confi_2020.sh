curl localhost:2019/config/apps/http/servers/hello/listen \
-X PATCH -H "Content-Type: application/json" -d '["localhost:2020-2025"]'

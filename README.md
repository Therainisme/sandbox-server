# sandbox-server

A secure sandbox that can compiles and runs CPP.

## Live demo

[Demo Site](https://notebook.therainisme.com/experment)

## Intro

The server use websocket to communicate on port 7777.

* source code

```cpp
#include <bits/stdc++.h>

using namespace std;

int main() {
    string input;
    cin >> input;
    cout << input;
    return 0;
}
```

* request

```json
{
    "text": "<source code>",
    "stdin":"Hello,World!"
}
```

* response

```json
{
    "memory": "<memory>",
    "time": "<memory>",
    "output": "<stdout>",
    "error":"<stderr>"
}
```

## Run server with docker

```shell
git clone https://github.com/Therainisme/sandbox-server
cd sandbox-server
```

Select a workspace folder and remember its absolute path.

And then modify the contents of the env file.

```shell
#example workspace=/home/welljuly/workspace/golang/sandbox/workspace

workspace=<your workspace path>
```

Finally

```shell
docker-compose up

# or 
# docker-compose up -d
```
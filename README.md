# Distributed Programming University Project 

# Dependencies

Since `client` within docker runs a GUI using `raylib`, we need to forward the desktop environment to docker.
So, a utility called `xhost` is needed (`xorg-xhost` package on arch).

Then run 

```bash
export DISPLAY=:0.0 &&\
xhost +local:docker &&\
docker compose up --build
```

# Deamon

[Setup](raspberry-pi/README.md)
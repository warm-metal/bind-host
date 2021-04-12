FROM bash:5

WORKDIR /root
COPY --from=docker.io/warmmetal/bind-host:latest /bind-host ./
ENV HOST_ROOTFS=""
ENV CRI_ADDR=""
ENV FSTAB=""
COPY *.sh ./
ENTRYPOINT ["./run.sh"]
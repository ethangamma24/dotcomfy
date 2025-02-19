FROM archlinux:latest

# Create a non-root user and group
RUN groupadd -r comfy && useradd -r -g comfy -d /home/comfy comfy

# Create and set up the work directory
RUN mkdir -p /home/comfy && chown -R comfy:comfy /home/comfy

RUN pacman -Syu && pacman -S --noconfirm sudo && usermod -aG wheel comfy
RUN usermod -aG wheel comfy

RUN echo 'comfy ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

RUN pacman -S --noconfirm which

# Switch to the non-root user
USER comfy
WORKDIR /home/comfy

# Copy the binary into the container
COPY --chown=comfy:comfy bin/dotcomfy bin/dotcomfy
COPY --chown=comfy:comfy tests/scripts/* tests/scripts/

# TODO: Copy test scenarios that are wrapped as bash scripts

# Default command (optional, replace with your binary execution command if needed)
# CMD ["bin/dotcomfy"]


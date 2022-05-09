# Start with nothing
FROM scratch

MAINTAINER jacobgadikian@gmail.com

# Add and decompress Arch Linux ARM rpi arm64 rootfs at /
ADD c2/ArchLinuxARM-odroid-c2-latest.tar.gz /

# Pacman Keyring
# Don't check disk space because we are in a container
RUN sed -i -e "s/^CheckSpace/#!!!CheckSpace/g" /etc/pacman.conf && \
		pacman-key --init && \
		pacman-key --populate archlinuxarm && \
		pacman-db-upgrade && \
		pacman -R --noconfirm openssh && \
		pacman -Syyu --noconfirm git sudo base-devel archlinux-keyring \
		ca-certificates \
		ca-certificates-mozilla \
		ca-certificates-utils \
		bash-completion \
		parted \
		e2fsprogs \
		dropbear \
		sudo \
		git \
		unbound \
		docker

# Add builduser and give wheel group sudos
#RUN useradd builduser -m && \
#	passwd -d builduser && \
#	printf 'builduser ALL=(ALL) ALL\n' | tee -a /etc/sudoers && \
#	echo '%wheel ALL=(ALL) ALL' >> /etc/sudoers.d/wheel

#USER builduser

# YAY AND HNSD
#RUN cd ~/ && \
#		git clone https://aur.archlinux.org/yay.git yay && \
#		cd yay && \
#		makepkg -si --noconfirm --clean --rmdeps && \
#		cd .. && \
#		rm -rf yay && \
#		cd ~/ && \
#		git clone https://github.com/faddat/hnsd-git && \
# 		cd hnsd-git && \
# 		makepkg -si --noconfirm --rmdeps --clean
# USER root




# First Boot service
COPY contrib/firstboot.sh /usr/local/bin/firstboot.sh
COPY contrib/firstboot.service /etc/systemd/system/firstboot.service

# Resize root fs
COPY contrib/resizerootfs /usr/sbin/resizerootfs
COPY contrib/resizerootfs.service /etc/systemd/system

# Copy DNS configuration so that stub resolver goes to hsd and falls back to GOOD(tm) public DNS
COPY contrib/dns /etc/systemd/resolved.conf

# HNSD Service: In testing hnsd has been unreliable
# COPY contrib/hnsd.service /etc/systemd/system/hnsd.service

# Greet Users Warmly
COPY contrib/motd /etc/


# Enable services
# Use the Pi's Hardware rng.  You may wish to modify depending on your needs and desires: https://wiki.archlinux.org/index.php/Random_number_generation#Alternatives
RUN systemctl enable firstboot && \
		systemctl enable resizerootfs && \
		systemctl enable systemd-resolved && \
# 		chmod +x /usr/bin/hnsd && \
		chmod +x /usr/local/bin/firstboot.sh && \
		chmod +x /usr/sbin/resizerootfs && \
#		systemctl enable hnsd && \
		# symlink systemd-resolved stub resolver to /etc/resolv/conf
		ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf && \
		systemctl enable dropbear && \
		echo "root:root" | chpasswd


# =================================================================
# CLEANUP: Make the OS new and shiny.
# =================================================================


# Remove cruft and let pacman check for free disk space
RUN rm -rf \
		/etc/*- \
		/var/lib/systemd/* \
		/var/lib/private/* \
		/var/log/* \
		/var/tmp/* \
		/tmp/* \
		/root/.bash_history \
		/root/.cache \
		/home/*/.bash_history \
		/home/*/.cache \
		`LC_ALL=C pacman -Qo /var/cache/* 2>&1 | grep 'error: No package owns' | awk '{print $5}'` && \
		sed -i -e "s/^#!!!CheckSpace/CheckSpace/g" /etc/pacman.conf

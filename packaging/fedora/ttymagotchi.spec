%global debug_package %{nil}
Name: ttymagotchi
Version: %{tty_version}
Release: 1%{?dist}
Summary: Animated virtual pets and interactive minigames in your terminal
License: MIT
URL: https://github.com/rene-timmermans/ttymagotchi
Source0: ttymagotchi-%{version}-linux-amd64.tar.gz
ExclusiveArch: x86_64
AutoReqProv: no

%description
Five species, ten growth stages, animated care and eight minigames.
Fork of Termagotchi by Eze Olea Figueroa.

%prep
%setup -q

%build
# The release archive contains a tested static Go binary.

%install
sh packaging/stage.sh %{buildroot}

%files
/usr/bin/ttymagotchi
/usr/share/applications/ttymagotchi.desktop
%license /usr/share/licenses/ttymagotchi/LICENSE
%doc /usr/share/doc/ttymagotchi/README.md

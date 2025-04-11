#
# spec file for joystick-monitor
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via
# https://github.com/mook/joystick-monitor/issues

Name:           joystick-monitor
Version:        0
Release:        0
Summary:        Inhibit screen saver during gamepad/joystick activity
License:        GPL-3.0-or-later
Group:          Hardware/Joystick
URL:            https://github.com/mook/joystick-monitor
Source0:        %{name}-%{version}.tar.zst
Source1:        vendor.tar.zst
Source2:        %{name}.preset
BuildRequires:  golang-packaging
# Needed for go -mod=vendor
BuildRequires:  golang(API) >= 1.24
BuildRequires:  zstd

%description
Monitors gamepads/joysticks used by applications and inhibits the screen saver
during activity

%prep
%autosetup -p1 -a1

%build
go build -mod=vendor -buildmode=pie
sed -i 's@ExecStart=/usr/local/bin/%{name}@ExecStart=%{_bindir}/%{name}@' %{name}.service

%install
%{__install} -D -m0755 %{name} %{buildroot}%{_bindir}/%{name}
%{__install} -D -m0644 %{name}.service %{buildroot}%{_userunitdir}/%{name}.service
%{__install} -D -m0644 %{SOURCE2} %{buildroot}%{_userpresetdir}/70-%{name}.preset
%{__mkdir_p} %{buildroot}%{_sbindir}
%{__ln_s} service %{buildroot}%{_sbindir}/rc%{name} # rpmlint/brp suse-missing-rclink

%pre
%service_add_pre %{name}.service

%post
%service_add_post %{name}.service

%preun
%service_del_preun %{name}.service

%postun
%service_del_postun %{name}.service

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_userunitdir}/%{name}.service
%{_sbindir}/rc%{name}
%{_userpresetdir}/70-%{name}.preset

%changelog

echo "Executing pre-installation tasks";

if getent group %{name} >/dev/null; then
  printf "INFO: %{name} group already exists\n"
else
  printf "INFO: %{name} group does not exist, creating ...\n"
  groupadd --system %{name}
fi

if getent passwd %{name} >/dev/null; then
  printf "INFO: %{name} user already exists\n"
else
  printf "INFO: %{name} group does not exist, creating ...\n"
  useradd --system -d /var/lib/%{name} -s /bin/bash -g %{name} %{name}
fi

echo "Completed pre-installation tasks";

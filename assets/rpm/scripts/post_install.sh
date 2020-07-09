echo "Executing post-installation tasks";
systemctl daemon-reload
mkdir -p /var/lib/%{name}
mkdir -p /var/lib/%{name}/.aws
touch /var/lib/%{name}/.aws/credentials
chown -R %{name}:%{name} /var/lib/%{name}
if test -f "/etc/%{name}/config.json"; then
    echo "Found existing configuration file"
else
    cp /etc/%{name}/config_template.json /etc/%{name}/config.json
fi
chown -R %{name}:%{name} /etc/%{name}
echo "Completed post-installation tasks";

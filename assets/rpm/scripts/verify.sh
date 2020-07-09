echo "Executing verification tasks";

/usr/local/bin/%{name} --version
if [ $? -eq 0 ]; then
    echo "Verification tasks were completed";
else
    echo "Verification tasks failed";
fi

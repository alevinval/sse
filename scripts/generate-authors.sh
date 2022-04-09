#/bin/bash

out="AUTHORS"

cat <<EOF > $out
# The author list of this library.
#
# This is the list of authors of the project for copyright purposes.
# When you contribute to the project you automatically become an author.
#
# Do not edit manually, this file is auto-generated.

EOF

git shortlog -se | awk -F '\t' '{ print $2 }' >> $out

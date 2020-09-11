#!/bin/bash
attribs=$(
gq https://pricing.infracost.io/graphql -q "
query {
    products (
    filter: {
      vendorName: \"aws\"
      region: \"us-east-1\"
      service: \"$1\"
      productFamily: \"$2\"
    }
  ){
    	productHash
    	attributes { key , value }
    }
}" | jq -r '.data.products[] | @base64'
)

N_files=$(echo $attribs | wc -w)

if [ $N_files -ne 1 ]; then
  echo -e "\n#####################################\n"
  echo Found $N_files different products
  echo -e "------------------\n"

  f=$(echo $attribs | cut -d' ' -f1 | base64 --decode | jq '.attributes ')
  s=$(echo $attribs | cut -d' ' -f2 | base64 --decode | jq '.attributes ')

  diff -U1 <( echo "$f" ) <( echo "$s" ) | grep -v "__typename" | grep -v "/dev/fd/"
else
  echo "Only 1 Match product found"
fi


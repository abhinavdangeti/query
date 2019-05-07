#!/bin/bash

echo Delete Buckets

Site=http://127.0.0.1:8091/pools/default/buckets/
Auth=Administrator:password
bucket=(customer orders product purchase review shellTest)


echo POST /pools/default/buckets

for i in "${bucket[@]}"
do
echo curl -u $Auth $Site$i
curl -X DELETE -u $Auth $Site$i
done

cd filestore

echo rm -rf data/
rm -rf data/

cd ../

echo Delete Users


UserSite=http://localhost:8091/settings/rbac/users/local/
for i in "${bucket[@]}"
do
Id=${i}owner
echo curl -X DELETE -u $Auth $UserSite$Id
curl -X DELETE -u $Auth $UserSite$Id
done


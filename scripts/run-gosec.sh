#!/bin/bash
   
   gosec -exclude-dir=sqlc \
         -exclude-dir=vendor \
         -exclude-generated \
         -exclude=G101,G102 \
         ./...
## knowledgebase

This script scans markdown documentation (the same input that is used by the Cosmos
backend) and uploads it to an Amazon Bedrock knowledgebase. The documents are split by
headers, and the header is included as metadata to be used as citations when the documents
are looked up.

The knowledgebase needs to be provisioned manually before running this script.

## Usage

Set up .env file using our current test knowledgebase.

```
AWS_PROFILE=...
AWS_DEFAULT_REGION=...
DATA_SOURCE_ID=4EROBRBALG
KNOWLEDGE_BASE_ID=CETGU0P5D7
```

Set up .venv

```sh
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

Run the script:

```sh
python create-kb.py
```

## Additional notes

The script can be used multiple times to "update" the knowledgebase. The documents are
keyed by a hash of folder and header. If the same document is uploaded again, it will
overwrite the previous version.

There is no delete functionality, so if documents are removed, they will be orphaned in
the knowledgebase.

The metadata attached to each document includes:
 - header: The markdown header for each document
 - folder: The folder the document was found in, excluding the "/docs/" path.
 
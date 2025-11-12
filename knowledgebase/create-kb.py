#!/usr/bin/env python3
import boto3, os, dotenv, re
import hashlib
dotenv.load_dotenv(override=True)
print(os.getcwd())
# Use the default AWS profile
session = boto3.Session()

# Create a client for the Bedrock Agent API
client = session.client('bedrock-agent')

# Example: List agents (replace with actual Bedrock Agent API call as needed)
response = client.list_agents()
print(response)

docs = []

def add_doc(folder_name: str, header_title: str, content_text: str) -> None:
    content_text = content_text.strip()
    if content_text == "": return
    if not header_title: return

    hash_text = f"{folder_name}__{header_title}"
    hash = hashlib.sha1(hash_text.encode()).hexdigest()[:10]

    docid = f"{folder_name}_{hash}"

    docs.append({
        "docid": docid,
        "folder_name": folder_name,
        "header_title": header_title,
        "content_text": content_text
    })

docs_dir = os.path.join(os.getcwd(), "docs")

for folder_name in os.listdir(docs_dir):
    folder_path = os.path.join(docs_dir, folder_name)
    if not os.path.isdir(folder_path):
        continue

    # Find files starting with a number, sort numerically
    files = [f for f in os.listdir(folder_path) if re.match(r'^\d+', f)]
    files.sort(key=lambda x: int(re.match(r'^(\d+)', x).group(1)))

    for filename in files:
        file_path = os.path.join(folder_path, filename)
        with open(file_path, "r") as f:
            lines = f.readlines()

        section_lines = []
        header_title = None
        for line in lines:
            header_match = re.match(r'^(#{1,6})\s+(.*)', line)
            if header_match:
                # Save previous section if exists
                if header_title and section_lines:
                    add_doc(folder_name, header_title, "".join(section_lines))
                    section_lines = []
                header_title = line.strip()
                section_lines.append(line)
            else:
                if header_title:
                    section_lines.append(line)
        # Save last section in file
        add_doc(folder_name, header_title, "".join(section_lines))

# Example: print the docs list
for doc in docs:
    print(f"Folder: {doc['folder_name']}, Header: {doc['header_title']}")
    print(doc['content_text'])
    print('-' * 40)

data_source_id = os.getenv("DATA_SOURCE_ID")
knowledge_base_id = os.getenv("KNOWLEDGE_BASE_ID")

if not data_source_id or not knowledge_base_id:
    print("DATA_SOURCE_ID and KNOWLEDGE_BASE_ID must be set in the environment.")
    exit(1)

aws_docs = []

for doc in docs:
    aws_docs.append({
        'metadata': {
            'type': 'IN_LINE_ATTRIBUTE',
            'inlineAttributes': [
                {
                    'key': 'header',
                    'value': {
                        'type': 'STRING',
                        'stringValue': doc['header_title'],
                    }
                },
                {
                    'key': 'folder',
                    'value': {
                        'type': 'STRING',
                        'stringValue': doc['folder_name'],
                    }
                },
            ],
        },
        'content': {
            'dataSourceType': 'CUSTOM',
            'custom': {
                'customDocumentIdentifier': {
                    'id': doc['docid']
                },
                'sourceType': 'IN_LINE',
                'inlineContent': {
                    'type': 'TEXT',
                    'textContent': {
                        'data': doc['content_text']
                    }
                }
            },
        }
    })

# Batch upload documents (max 10 per request)
for i in range(0, len(aws_docs), 10):
    batch = aws_docs[i:i+10]
    try:
        response = client.ingest_knowledge_base_documents(
            knowledgeBaseId=knowledge_base_id,
            dataSourceId=data_source_id,
            documents=batch
        )
    except Exception as e:
        print(f"Error uploading batch {i//10 + 1}: {e}")
        continue
    print(f"Uploaded batch {i//10 + 1}: {response}")

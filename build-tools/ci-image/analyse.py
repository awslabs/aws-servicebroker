#!/usr/bin/env python3
from github import Github
import os


TOKEN = os.environ['GITHUB_TOKEN']
ORG = os.environ['GITHUB_ORG']
REPO = os.environ['GITHUB_REPO']
HEAD = os.environ['CODEBUILD_RESOLVED_SOURCE_VERSION']

print("Comparing %s with master for repo %s/%s" % (HEAD, ORG, REPO))
g = Github(TOKEN)
g.get_user()
r = g.get_repo(ORG + "/" + REPO, lazy=False)

BASE = r.get_branch('master').commit.sha

files = r.compare(BASE, HEAD).files

all_apbs = [d for d in os.listdir('./templates/') if os.path.isdir('./templates/' + d)]
apbs = []
build_all = False
broker = False
for f in files:
    if f.filename.startswith('templates/'):
        name = f.filename.split('/')[1]
        if name not in apbs:
            apbs.append(name)
    elif f.filename == 'broker_image_sha':
        broker = True
    elif f.filename.startswith('build-tools/apb-packaging/sb_cfn_package/'):
        build_all = True

with open('apbs', 'w') as f:
    if build_all:
        for apb in all_apbs:
            f.write('%s\n' % apb)
    else:
        for apb in apbs:
            f.write('%s\n' % apb)

with open('broker', 'w') as f:
    f.write('%s' % str(broker))


#!/usr/bin/env python

import os
import jinja2
import yaml

dir_path = os.path.dirname(os.path.realpath(__file__))

def render(tpl_path, content):
    abs_path = os.path.join(dir_path, tpl_path)
    path, filename = os.path.split(abs_path)
    return jinja2.Environment(
        loader=jinja2.FileSystemLoader(path or './')
    ).get_template(filename).render(content)

with open(os.path.join(dir_path, 'k8s-variables.yaml'), 'r') as content_file:
    data = content_file.read()

content = yaml.load(data)

result = render('./k8s-aws-service-broker.yaml.j2', content)

with open(os.path.join(dir_path, 'k8s-aws-service-broker.yaml'), 'w') as rendered_file:
    rendered_file.write(result)

#!/bin/env python3
''' node-balance-compare.py

print out some information about nodes to help in debugging cluster
autoscaler balance similar nodes failures.

usage:
    oc get nodes -o json | node-balance-compare.py > nodes.html
'''
import fileinput
import io
import json
import sys


class Node:
    def __init__(self, data):
        self._data = data

    def capacity(self):
        return self._data.get('status', {}).get('capacity', {})

    def is_master(self):
        return self._data.get('metadata', {}).get('labels', {}).get('node-role.kubernetes.io/master') is not None

    def labels(self):
        return self._data.get('metadata', {}).get('labels', {})

    def machine(self):
        return self._data.get('metadata', {}).get('annotations', {}).get('machine.openshift.io/machine', '')

    def name(self):
        return self._data.get('metadata', {}).get('name', 'no name specified')


def main():
    buffer = io.StringIO()

    for line in fileinput.input(encoding="utf-8"):
        buffer.write(line)

    buffer.seek(0)

    items = json.load(buffer)
    items = items.get('items', [])
    nodes = [Node(i) for i in items]
    nodes = filter(lambda n: not n.is_master(), nodes)
    nodes = sorted(nodes, key=lambda n: n.name())

    namerow = []
    machinerow = []
    labelsrow = []
    capacityrow = []
    for n in nodes:
        namerow.append(n.name())
        machinerow.append(n.machine())
        labelsrow.append(n.labels())
        capacityrow.append(n.capacity())

    tablerows = '<tr><th>name</th>'
    for c in namerow:
        tablerows += f'<td>{c}</td>'
    tablerows += '</tr>'

    tablerows += '<tr><th>machine</th>'
    for c in machinerow:
        tablerows += f'<td>{c}</td>'
    tablerows += '</tr>'

    tablerows += '<tr><th>labels</th>'
    for c in labelsrow:
        keys = sorted(c.keys())
        l = ''
        for k in keys:
            l += f'{k}: {c[k]}<br/>'

        tablerows += f'<td>{l}</td>'
    tablerows += '</tr>'

    tablerows += '<tr><th>capacity</th>'
    for c in capacityrow:
        keys = sorted(c.keys())
        l = ''
        for k in keys:
            l += f'{k}: {c[k]}<br/>'

        tablerows += f'<td>{l}</td>'
    tablerows += '</tr>'

    html = '''<!doctype html>
<html>
  <body>
    <table>
        {tablerows}
    </table>
  </body>
</html>
    '''
    print(html.format(tablerows=tablerows))


if __name__ == '__main__':
    main()

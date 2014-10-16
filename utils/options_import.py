#!/usr/bin/env python3
# -*- coding: utf-8 -*-
#
# This file is part of agora-api.
# Copyright (C) 2014 Eduardo Robles Elvira <edulix AT agoravoting DOT com>

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

import json
import copy
import sys
import codecs

def read_csv_to_dicts(path, sep=";"):
    '''
    Given a file in CSV format, convert it to a dictionary.
    Example file (example.csv):

    ID,Name,Comment
    1,Fulanito de tal,Nothing of interest
    2,John Doe,

    This file would be converted into the following if you call to
    read_csv_to_dicts("example.csv", sep=","):
    {
            "id": 2,
            "media_url":"",
            "details_title":"",
            "details":"",
            "value":"Resolución sobre legislación - Ley de Voto Electrónico (e-voting) ",
            "category": "Nuevas tecnologías",
            "a":"ballot/answer",
            "urls":[
              {"title": "Ver", "url": "https://web-podemos.s3.amazonaws.com/wordpress/wp-content/uploads/2014/10/REsolucionVotoElectronico.pdf"}
            ]
    }
    '''
    ret = []
    n = -1
    m = 0
    baseopt = {
      "id": 1,
      "media_url":"",
      "details_title":"",
      "details":"",
      "value":"",
      "category": "",
      "a":"ballot/answer",
      "urls":[
        {"title": "Ver", "url": ""}
      ]
    }
    with open(path, mode='r', encoding="utf-8", errors='strict') as f:
        for line in f:
            n += 1
            if n == 0:
              continue
            line = line.rstrip()
            item = copy.deepcopy(baseopt)
            values = line.split(sep)
            if len(values) == 3:
              m += 1
              item['id'] = m
              item['value'] = values[1].strip()
              item['category'] = values[2].strip()
              item['urls'][0]['url'] = values[0].strip()
              ret.append(item)
    return ret

if __name__ == '__main__':
    data = read_csv_to_dicts(sys.argv[1], '\t')
    print(json.dumps(data, indent=4, ensure_ascii=False))


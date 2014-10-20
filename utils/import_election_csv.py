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
import re

def read_csv_to_dicts(path):
    '''
    Given a file in CSV format, convert it to a dictionary.
    Example file (example.csv):

    Título de votación:,Votación de borradores
    Descripción,Descripción de la votación lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum
    URL,TÍTULO,FIRMANTES,DOCUMENTO
    https://whatever.com,Propuesta de Pablo,"Firmante 1, Firmante 2..",Documento ético
    https://whatever.com,Propuesta de Pablo,"Firmante 1, Firmante 2..",Documento político
    https://whatever.com,Propuesta de Pablo,"Firmante 1, Firmante 2..",Documento organizativo

    This file would be converted into the following

    {
        "pretty_name": "Votación de borradores",
        "voting_start_date": "",
        "min": 0,
        "max": 1,
        "is_recurring": false,
        "director": "wadobo-auth1",
        "id": 1,
        "description": "Descripción de la votación lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum",
        "url": "http://podemos.info",
        "questions_data": [
            {
                "question": "Documento organizativo",
                "description": "",
                "tally_type": "APPROVAL",
                "min": 0,
                "a": "ballot/question",
                "max": 1,
                "layout": "drafts-election",
                "randomize_answer_order": false,
                "answers": [
                    {
                        "media_url": "",
                        "details_title": "",
                        "value": "Propuesta de Pablo",
                        "category": "",
                        "a": "ballot/answer",
                        "details": "Firmante 1, Firmante 2..",
                        "urls": [
                            {
                                "url": "https://whatever.com",
                                "title": "Ver"
                            }
                        ],
                        "id": 3
                    }
                ],
                "num_winners": 1
            },
            {
                "question": "Documento político",
                "description": "",
                "tally_type": "APPROVAL",
                "min": 0,
                "a": "ballot/question",
                "max": 1,
                "layout": "drafts-election",
                "randomize_answer_order": false,
                "answers": [
                    {
                        "media_url": "",
                        "details_title": "",
                        "value": "Propuesta de Pablo",
                        "category": "",
                        "a": "ballot/answer",
                        "details": "Firmante 1, Firmante 2..",
                        "urls": [
                            {
                                "url": "https://whatever.com",
                                "title": "Ver"
                            }
                        ],
                        "id": 2
                    }
                ],
                "num_winners": 1
            },
            {
                "question": "Documento ético",
                "description": "",
                "tally_type": "APPROVAL",
                "min": 0,
                "a": "ballot/question",
                "max": 1,
                "layout": "drafts-election",
                "randomize_answer_order": false,
                "answers": [
                    {
                        "media_url": "",
                        "details_title": "",
                        "value": "Propuesta de Pablo",
                        "category": "",
                        "a": "ballot/answer",
                        "details": "Firmante 1, Firmante 2..",
                        "urls": [
                            {
                                "url": "https://whatever.com",
                                "title": "Ver"
                            }
                        ],
                        "id": 1
                    }
                ],
                "num_winners": 1
            }
        ],
        "extra": [],
        "authorities": "openkratio-authority",
        "layout": "drafts-election",
        "voting_end_date": "",
        "hash": "",
        "title": ""
    }

    '''


    ret = {
      "id": 1,
      "hash": "",
      "pretty_name":"",
      "description":"",
      "title":"",
      "layout": "drafts-election",
      "voting_start_date": "",
      "voting_end_date": "",
      "is_recurring": False,
      "director": "",
      "authorities": "",
      "url": "",
      "extra": [],
      "min": 0,
      "max": 1,
    }
    basequestion = {
      "a":"ballot/question",
      "question":"",
      "description": "",
      "layout":"drafts-election",
      "tally_type":"APPROVAL",
      "randomize_answer_order":False,
      "min":0,
      "max":1,
      "num_winners":1,
      "answers":[]
    }
    baseopt = {
      "id": 1,
      "media_url":"",
      "details_title":"",
      "isPack": False,
      "details":"",
      "value":"",
      "category": "",
      "a":"ballot/answer",
      "urls":[
        {"title": "Ver", "url": ""}
      ]
    }
    questions = {}
    n = -1
    m = 0

    with open(path, mode='r', encoding="utf-8", errors='strict') as f:
        for line in f:
            n += 1
            if n == 0:
                values = line.split(',')
                ret['pretty_name'] = values[1].strip()
                continue
            elif n == 1:
                values = line.split(',')
                ret['description'] = values[1].strip()
                pass
            elif n == 2:
                # skip headers for question data
                continue
            else:
                # print(n, line)
                line = line.rstrip()
                values = re.split(''',(?=(?:[^'"]|'[^']*'|"[^"]*")*$)''', line)
                item = copy.deepcopy(baseopt)
                question = values[3].strip()

                if len(values) == 5:
                    m += 1
                    item['id'] = m
                    item['value'] = values[1].strip()
                    item['value'] = values[1].strip()
                    item['details'] = values[2].strip().replace('"', '')
                    item['urls'][0]['url'] = values[0].strip()
                    divisible = values[4].strip().lower()
                    item['isPack'] = divisible == "no"

                    if question not in questions:
                        questions[question] = copy.deepcopy(basequestion)
                        questions[question]['question'] = question

                    questions[question]['answers'].append(item)


        ret['questions_data'] = list(questions.values())

    return ret

if __name__ == '__main__':
    data = read_csv_to_dicts(sys.argv[1])
    print(json.dumps(data, indent=4, ensure_ascii=False, sort_keys=True))

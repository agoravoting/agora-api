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
    https://whatever.com,Propuesta de Pablo,"Firmante 1, Firmante 2..",Documento ético,no
    https://whatever.com,Propuesta de Pablo,"Firmante 1, Firmante 2..",Documento político,no
    https://whatever.com,Propuesta de Pablo,"Firmante 1, Firmante 2..",Documento organizativo,si

    This file would be converted into the following

    {
        "authorities": "",
        "description": "Descripción de la votación lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum
    lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum",
        "director": "",
        "extra": [],
        "hash": "",
        "id": 1,
        "is_recurring": false,
        "layout": "drafts-election",
        "max": 1,
        "min": 0,
        "pretty_name": "Votación de borradores",
        "questions_data": [
            {
                "a": "ballot/question",
                "answers": [
                    {
                        "a": "ballot/answer",
                        "category": "",
                        "details": "Firmante 1, Firmante 2..",
                        "details_title": "",
                        "id": 2,
                        "isPack": true,
                        "media_url": "",
                        "urls": [
                            {
                                "title": "Ver",
                                "url": "https://whatever.com"
                            }
                        ],
                        "value": "Propuesta de Pablo"
                    }
                ],
                "description": "",
                "layout": "drafts-election",
                "max": 1,
                "min": 0,
                "num_seats": 1,
                "question": "Documento político",
                "randomize_answer_order": false,
                "tally_type": "APPROVAL"
            },
            {
                "a": "ballot/question",
                "answers": [
                    {
                        "a": "ballot/answer",
                        "category": "",
                        "details": "Firmante 1, Firmante 2..",
                        "details_title": "",
                        "id": 1,
                        "isPack": true,
                        "media_url": "",
                        "urls": [
                            {
                                "title": "Ver",
                                "url": "https://whatever.com"
                            }
                        ],
                        "value": "Propuesta de Pablo"
                    }
                ],
                "description": "",
                "layout": "drafts-election",
                "max": 1,
                "min": 0,
                "num_seats": 1,
                "question": "Documento ético",
                "randomize_answer_order": false,
                "tally_type": "APPROVAL"
            },
            {
                "a": "ballot/question",
                "answers": [
                    {
                        "a": "ballot/answer",
                        "category": "",
                        "details": "Firmante 1, Firmante 2..",
                        "details_title": "",
                        "id": 3,
                        "isPack": false,
                        "media_url": "",
                        "urls": [
                            {
                                "title": "Ver",
                                "url": "https://whatever.com"
                            }
                        ],
                        "value": "Propuesta de Pablo"
                    }
                ],
                "description": "",
                "layout": "drafts-election",
                "max": 1,
                "min": 0,
                "num_seats": 1,
                "question": "Documento organizativo",
                "randomize_answer_order": false,
                "tally_type": "APPROVAL"
            }
        ],
        "title": "",
        "url": "",
        "voting_end_date": "",
        "voting_start_date": ""
    }

    '''


    ret = {
      "id": 10205,
      "hash": "",
      "pretty_name":"",
      "description":"",
      "title":"votacion-de-borradores",
      "layout": "drafts-election",
      "voting_start_date": "2013-12-06T18:17:14.457000",
      "voting_end_date": "2013-12-09T18:17:14.457000",
      "is_recurring": False,
      "director": "wadobo-auth1",
      "authorities": "openkratio-authority",
      "url": "http://podemos.info",
      "extra": [],
      "min": 0,
      "max": 3,
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
      "num_seats":1,
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
    sorted_keys = []
    n = -1

    with open(path, mode='r', encoding="utf-8", errors='strict') as f:
        for line in f:
            n += 1
            if n == 0:
                values = line.split('\t')
                ret['pretty_name'] = values[1].strip()
                continue
            elif n == 1:
                values = line.split('\t')
                ret['description'] = values[1].strip()
                pass
            elif n == 2:
                # skip headers for question data
                continue
            else:
                # print(n, line)
                line = line.rstrip()
                values = line.split('\t')
                item = copy.deepcopy(baseopt)
                question = values[3].strip()

                if len(values) >= 4:
                    item['sort_order'] = n
                    item['value'] = values[1].strip()
                    item['details'] = values[2].strip().replace('"', '')
                    item['urls'][0]['url'] = values[0].strip()
                    divisible = (len(values) > 4 and values[4].strip().lower())
                    item['isPack'] = divisible == "no"

                    if question not in questions:
                        sorted_keys.append(question)
                        questions[question] = copy.deepcopy(basequestion)
                        questions[question]['question'] = question

                    item['id'] = 1 + len(questions[question]['answers'])
                    questions[question]['answers'].append(item)


        ret['questions_data'] = list([questions[q] for q in sorted_keys])

    return ret

if __name__ == '__main__':
    data = read_csv_to_dicts(sys.argv[1])
    print(json.dumps(data, indent=4, ensure_ascii=False, sort_keys=True))

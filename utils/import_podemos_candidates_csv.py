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

def read_csv_to_dicts(path, eid):
    ret = {
        "id": eid,
        "pretty_name": "Votación de candidatos",
        "description": "Selecciona los candidatos a los Órganos Internos de Podemos, para la Secretaría General, el Consejo Ciudadano y la Comisión de Garantía Democráticas",
        "director": "wadobo-auth1",
        "authorities": "openkratio-authority",
        "extra": [],
        "hash": "",
        "is_recurring": False,
        "layout": "pcandidates-election",
        "questions_data": [],
        "title": "votacion-de-candidatos",
        "url": "http://podemos.info",
        "voting_end_date": "2014-11-10T09:00:00.000000",
        "voting_start_date": "2014-11-15T00:00:00.000000"
    }

    base_question = {
        "a": "ballot/question",
        "question_id": 0,
        "description": "",
        "layout": "pcandidates-election",
        "max": 1,
        "min": 0,
        "num_seats": 1,
        "question": "Secretaría General",
        "question_slug": "secretario",
        "randomize_answer_order": True,
        "tally_type": "APPROVAL",
        "answers": []
    }

    base_answer = {
        "id": 1,
        "a": "ballot/answer",
        "category": "Equipo de Enfermeras",
        "details": "",
        "details_title": "",
        "media_url": "",
        "sort_order": 1,
        "urls": [],
        "value": "Fulanita de tal"
    }
    questions = {
      'secretario': copy.deepcopy(base_question),
      'consejo': copy.deepcopy(base_question),
      'garantias': copy.deepcopy(base_question)
    }

    questions['secretario']['question_id'] = 0
    questions['secretario']['question'] = "Secretaría General"
    questions['secretario']['question_slug'] = 'secretario'
    questions['secretario']['num_seats'] = 1
    questions['secretario']['max'] = 1

    questions['consejo']['question_id'] = 1
    questions['consejo']['question'] = "Consejo Ciudadano"
    questions['consejo']['question_slug'] = 'consejo'
    questions['consejo']['num_seats'] = 62
    questions['consejo']['max'] = 62

    questions['garantias']['question_id'] = 2
    questions['garantias']['question'] = "Comisión de Garantías"
    questions['garantias']['question_slug'] = 'garantias'
    questions['garantias']['num_seats'] = 10
    questions['garantias']['max'] = 10

    linenum = -1 # linenum

    with open(path, mode='r', encoding="utf-8", errors='strict') as f:
        for line in f:
            linenum += 1
            if linenum == 0:
                continue

            line = line.rstrip()
            values = line.split('\t')
            item = copy.deepcopy(base_answer)
            question_slug = values[0].strip()

            if values[0].strip().startswith("Secretaría"):
                question_slug = "secretario"
            elif values[0].strip().startswith("Consejo"):
                question_slug = "consejo"
            elif values[0].strip().startswith("Comisión"):
                question_slug = "garantias"

            if len(values) >= 3:
                item['value'] = values[2].strip().replace('"', '\'')
                item['category'] = values[1].strip()

                item['id'] = 1 + len(questions[question_slug]['answers'])
                item['sort_order'] = item['id']
                questions[question_slug]['answers'].append(item)
            else:
                print("!!! error parsing linenum: %d, line: %s.." % (linenum, line[:40]))
                exit(1)

    sorted_keys = ['secretario', 'consejo', 'garantias']
    ret['questions_data'] = list([questions[q] for q in sorted_keys])
    return ret

if __name__ == '__main__':
    data = read_csv_to_dicts(sys.argv[1], sys.argv[2])
    print(json.dumps(data,
        indent=4, ensure_ascii=False, sort_keys=True, separators=(',', ': ')))

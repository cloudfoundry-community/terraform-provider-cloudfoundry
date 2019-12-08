#!/usr/bin/env python3
"""This script helps to migrate the tf files and the tfstates files to support the terraform cloud foundry provider form 0.9.9"""

import sys
import re
import shutil
import os
import json
from argparse import ArgumentParser


class Converter:
    def __init__(self, tf_path: str, debug: bool, do_backup: bool):
        self.debug = debug
        self.do_backup = do_backup
        self.tf_path = tf_path

    def convert_cf2cloudfoundry(self, obj):
        raise NotImplementedError("Not implemented, use one of the child classes")


class JSONConverter(Converter):
    def migrate(self):
        # Step 0 make backup
        if self.do_backup:
            shutil.copy(self.tf_path, self.tf_path + ".backup")
        with open(self.tf_path, "r") as state_file:
            state_dict = json.load(state_file)
            new_state_file = open(self.tf_path + ".new", "w")
            # Step 1. rename every cf_ to cloudfoundry_
            self.convert_cf2cloudfoundry(state_dict)
            # Step 2. remove fork specific features
            # Step 2a. remove recursive
            self.remove_recursive_delete(state_dict)
            # Step 2b. remove disable_blue_green_deployment
            self.remove_disable_blue_green_deployment(state_dict)
            # Step 3. add new flags with default values if any - not needed
            json.dump(obj=state_dict, fp=new_state_file, indent=4)
            new_state_file.close()
        os.remove(self.tf_path)
        shutil.move(self.tf_path + ".new", self.tf_path)

    def convert_cf2cloudfoundry(self, state_dict: dict):
        cf_prefix_regex = r'^cf_'
        cf_prefix_regex_replacement = "cloudfoundry_"
        cf_regexes = {
            cf_prefix_regex: cf_prefix_regex_replacement,
            r'^data.cf_': "data.cloudfoundry_"
        }
        provider_regex = r'^provider.cf'
        provider_regex_replacement = r'provider.cloudfoundry'
        modules = state_dict['modules']
        for module in modules:
            for key, value_dict in list(module['resources'].items()):
                # 1. change the stuff in the value dict
                # 1a. change each moduel in depends_on
                depending_modules = value_dict['depends_on']
                new_depending_modules = list()
                for depending_module in depending_modules:
                    regex_matched = False
                    for regex, replacement in cf_regexes.items():
                        cf_regex = re.compile(regex)
                        if cf_regex.search(depending_module):
                            new_depending_module = cf_regex.sub(replacement, depending_module)
                            if self.debug:
                                print(f"Key: \"{key}\"->depends_on adding {new_depending_module} as replaced {depending_module}")
                            # add new module, and break out the for cycle as one of the regexes matches
                            new_depending_modules.append(new_depending_module)
                            # exit if regex_matched
                            regex_matched = True
                            break
                    if not regex_matched:
                        if self.debug:
                            print(f"Key: \"{key}\"->depends_on adding back/skipping {depending_module}")
                        new_depending_modules.append(depending_module)
                value_dict['depends_on'] = new_depending_modules

                # 1b. change type
                cf_prefix_re = re.compile(cf_prefix_regex)
                if cf_prefix_re.search(value_dict['type']):
                    value_dict['type'] = cf_prefix_re.sub(cf_prefix_regex_replacement, value_dict['type'])
                    if self.debug:
                        print(f"Key: \"{key}->type\" replaced to {value_dict['type']} as matched {cf_prefix_regex}")

                # 1c. change provider
                provider_re = re.compile(provider_regex)
                if provider_re.search(value_dict['provider']):
                    value_dict['provider'] = provider_re.sub(provider_regex_replacement, value_dict['provider'])
                    if self.debug:
                        print(
                            f"Key: \"{key}->provider\" replaced to {value_dict['provider']} as matched {provider_regex}")

                # 2. replace the key itself
                for regex, replacement in cf_regexes.items():
                    cf_regex = re.compile(regex)
                    if cf_regex.search(key):
                        new_key = cf_regex.sub(replacement, key)
                        if self.debug:
                            print(f"Key: \"{key}\" matches {regex}, replacing match with {new_key}")
                        module['resources'][new_key] = module['resources'].pop(key)

    def remove_recursive_delete(self, state_dict: dict):
        self.__remove_attribute__(state_dict, "recursive_delete")

    def remove_disable_blue_green_deployment(self, state_dict: dict):
        self.__remove_attribute__(state_dict, "disable_blue_green_deployment")

    def __remove_attribute__(self, state_dict: dict, attribute: str):
        modules = state_dict['modules']
        for module in modules:
            for key, value_dict in module['resources'].items():
                if attribute in value_dict['primary']['attributes']:
                    if self.debug:
                        print(f"In: \"{key}\" deleting \"{attribute}\" attribute")
                    value_dict['primary']['attributes'].pop(attribute)


class HCLConverter(Converter):
    def migrate(self):
        # Step 0 make backup
        if self.do_backup:
            shutil.copy(self.tf_path, self.tf_path + ".backup")
        with open(self.tf_path, "r") as tf_file:
            new_tf_file = open(self.tf_path + ".new", "w")
            for line in tf_file:
                # Step 1. rename every cf_ to cloudfoundry_
                line_1 = self.convert_cf2cloudfoundry(line)
                # Step 2. remove fork specific features
                # Step 2a. remove disable_blue_green_deployment
                line_2 = self.remove_disable_blue_green_deployment(line_1)
                # Step 2b. remove recursive
                new_tf_file.write(self.remove_recursive_delete(line_2))
                # Step 3. add new flags with default values if any - not needed
            new_tf_file.close()
        os.remove(self.tf_path)
        shutil.move(self.tf_path + ".new", self.tf_path)

    def convert_cf2cloudfoundry(self, line: str):
        cf_regexes = {
            r'^provider "cf"': 'provider "cloudfoundry"',  # Adjust provider
            r'^resource "cf_': 'resource "cloudfoundry_',  # Adjust resource
            r'^data "cf_': 'data "cloudfoundry_',  # Adjust data
            r'\$\{data.cf_': '${data.cloudfoundry_',  # Adjust data. in variable
            r'\$\{cf_': '${cloudfoundry_',  # Adjust cf_ ...in variable
            r'\["cf_': '["cloudfoundry_',  # Adjust mostly depends only
            r',\s*cf_': ', cloudfoundry_'  # Adjust joins etc
        }
        for regex, replacement in cf_regexes.items():
            cf_regex = re.compile(regex)
            if cf_regex.search(line):
                if self.debug:
                    print(f"Line: \"{line[:-1]}\" matches {regex}, replacing match with {replacement}")
                return cf_regex.sub(replacement, line)
        return line

    def remove_recursive_delete(self, line: str):
        recursive_delete_regex = re.compile(r'^\s*recursive_delete\s*=\s*\w+\s*')
        if recursive_delete_regex.match(line):
            if self.debug:
                print(f"Removing \"recursive_delete\" attribute")
            return ''
        return line

    def remove_disable_blue_green_deployment(self, line: str):
        disable_blue_green_deployment_regex = re.compile(r'^\s*disable_blue_green_deployment\s*=\s*\w+\s*')
        if disable_blue_green_deployment_regex.match(line):
            if self.debug:
                print(f"Removing \"disable_blue_green_deployment\" attribute")
            return ''
        return line


if __name__ == "__main__":
    if sys.version_info < (3, 6):
        sys.exit('Sorry, Python < 3.6 is not supported')
    parser = ArgumentParser("Terraform config and state file migration script for the cloud foundry provider")
    parser.add_argument('-t', '--type', dest='type', action="store", default=None, choices=("tf", "state"),
                        required=True, help="Type of the migration, accepted values: state, tf")
    parser.add_argument('-n', '--name', dest='tf_path', action="store", default=None, required=True,
                        help="Name of the file")
    parser.add_argument('-b', '--skip-backup', dest='do_backup', action="store_false", default=True,
                        help="Skips backing up the file")
    parser.add_argument('-d', '--debug', dest='debug', action="store_true", default=False,
                        help="Debug mode - prints a lot of useful information")
    args = parser.parse_args()

    if args.type == "tf":
        converter = HCLConverter(args.tf_path, args.debug, args.do_backup)
    elif args.type == "state":
        converter = JSONConverter(args.tf_path, args.debug, args.do_backup)
    converter.migrate()

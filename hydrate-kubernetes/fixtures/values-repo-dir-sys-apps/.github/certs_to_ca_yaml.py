#!/usr/bin/env python3

import glob
import argparse
import yaml


def configure_parser():
    parser = argparse.ArgumentParser(prog="ca_to_yaml", add_help=True)
    parser.add_argument(
        "--ca_certs_path",
        help="Path to ca_certs",
        required=True,
        type=str,
    )
    parser.add_argument(
        "--ca_yml_path",
        help="Path where to put the ca.yml file",
        required=True,
        type=str,
    )
    return parser

if __name__ == "__main__":
    parser = configure_parser()
    args = vars(parser.parse_args())

    full_ca = ''

    for file in sorted(glob.glob(f"{args['ca_certs_path']}/*.crt")):
        with open(file, encoding="utf-8") as f:
            full_ca += f.read()

    dict_file = {'councilbox-server' : {'ca_secret' : {'crts' :  full_ca }}}
    #print(full_ca)


    with open(f"{args['ca_yml_path']}", 'w', encoding="utf-8") as yamlfile:
        yaml.dump(dict_file, yamlfile, default_style = '|')

    print(yaml.dump(dict_file, default_style = '|'))


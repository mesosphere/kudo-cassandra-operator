#!/usr/bin/env python3

import argparse
import logging
import os
import os.path as path
import sys
import yaml
from pytablewriter import MarkdownTableWriter
from pytablewriter.style import Style

SCRIPT_DIRECTORY = path.dirname(path.realpath(__file__))

DEFAULT_PARAMS_YAML = path.realpath(
    path.join(SCRIPT_DIRECTORY, "../operator/params.yaml")
)

DEFAULT_DOCS_MARKDOWN = path.realpath(
    path.join(SCRIPT_DIRECTORY, "../docs/parameters.md")
)

log = logging.getLogger(__name__)


def generate_markdown_table(input_params_yaml, output_markdown_file_path):
    try:
        with open(input_params_yaml, "r") as input:
            yamlData = yaml.safe_load(input)

            groups = yamlData["groups"]
            parameters = yamlData["parameters"]

            groupTable = MarkdownTableWriter()
            groupTable.table_name = "Groups"
            groupTable.styles = [
                Style(align="left"),
                Style(align="left"),
            ]
            groupTable.headers = [ "Group", "Description" ] #"[field.capitalize() for field in group_fields]
            groupTable.margin = 1
            groupTable.value_matrix = [
                # list([group.get(field) for field in group_fields])
                gen_group_header(group) for group in groups
            ]

            emptyGroup = {
                'name': '',
                'displayName': 'Ungrouped Parameters',
                'description': 'All parameters that are not assigned to a specific group.'
            }

            if output_markdown_file_path:
                try:
                    with open(output_markdown_file_path, "w") as output:
                        groupTable.stream = output
                        groupTable.write_table()

                        for group in groups:
                            gen_group(group, parameters, output)

                        gen_group(emptyGroup, parameters, output)

                except Exception as e:
                    logging.error(
                        f"Failed to output Markdown to {output_markdown_file_path}",
                        e,
                    )
                    return 1
            else:
                logging.error(
                    f"Required markdown output"
                )

            return 0
    except Exception as e:
        logging.error(
            f"Failed to generate Markdown from {input_params_yaml}", e
        )
        return 1


def gen_group_header(group):
    display_name = group.get("displayName", group.get("name"))

    return [f'[{display_name}](#{group.get("name")})', group.get("description")]

def gen_group(group, parameters, output):
    parameter_fields = ["name", "description", "default"]

    display_name = group.get("displayName", group.get("name"))

    output.write(f'## <a name="{group.get("name")}"></a>  {display_name} \n')
    output.write(f'Name: {group.get("name")}  \n')
    if "description" in group:
        output.write(f'{group.get("description")}\n')

    output.write('\n\n')

    paramTable = MarkdownTableWriter()

    # paramTable.table_name = group["displayName"]
    paramTable.styles = [
        Style(align="left", font_weight="bold"),
        Style(align="left"),
        Style(align="left"),
    ]
    paramTable.headers = [field.capitalize() for field in parameter_fields]
    paramTable.margin = 1

    paramTable.value_matrix = [
        list([parameter.get(field) for field in parameter_fields])
        for parameter in parameters if parameter.get("group", "") == group.get("name")
    ]

    paramTable.stream = output
    paramTable.write_table()

    return 0


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Generate a markdown table of all parameters in a KUDO "
                    + "Operator params.yaml file"
    )

    parser.add_argument(
        "--input-params-yaml",
        help=f"params.yaml input file. Defaults to {DEFAULT_PARAMS_YAML}",
        default=DEFAULT_PARAMS_YAML,
    )
    parser.add_argument(
        "--output-markdown-file",
        help="Relative path for output file with markdown table",
        default=DEFAULT_DOCS_MARKDOWN,
    )

    args = parser.parse_args()

    input_params_yaml_path = path.realpath(
        path.join(os.getcwd(), args.input_params_yaml)
    )

    output_markdown_file_path = (
        path.realpath(path.join(os.getcwd(), args.output_markdown_file))
        if args.output_markdown_file
        else None
    )

    return generate_markdown_table(
        input_params_yaml_path, output_markdown_file_path
    )


if __name__ == "__main__":
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s",
        datefmt="%Y-%m-%d %H:%M:%SZ",
    )
    sys.exit(main())

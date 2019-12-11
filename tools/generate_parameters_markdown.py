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

log = logging.getLogger(__name__)


def generate_markdown_table(input_params_yaml, output_markdown_file_path):
    try:
        with open(input_params_yaml, "r") as input:
            parameters = yaml.safe_load(input)["parameters"]
            parameter_fields = ["name", "description", "default"]

            writer = MarkdownTableWriter()
            writer.table_name = "Parameters"
            writer.styles = [
                Style(align="left", font_weight="bold"),
                Style(align="left"),
                Style(align="left"),
            ]
            writer.headers = [field.capitalize() for field in parameter_fields]
            writer.margin = 1

            writer.value_matrix = [
                list([parameter.get(field) for field in parameter_fields])
                for parameter in parameters
            ]

            if output_markdown_file_path:
                try:
                    with open(output_markdown_file_path, "w") as output:
                        writer.stream = output
                        writer.write_table()
                except Exception as e:
                    logging.error(
                        f"Failed to output Markdown to {output_markdown_file_path}",
                        e,
                    )
                    return 1
            else:
                print(writer.dumps())

            return 0
    except Exception as e:
        logging.error(
            f"Failed to generate Markdown from {input_params_yaml}", e
        )
        return 1


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

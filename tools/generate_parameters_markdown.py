#!/usr/bin/env python3

import argparse
import logging
import os.path as path
import sys
import yaml
from pytablewriter import MarkdownTableWriter
from pytablewriter.style import Style


SCRIPT_DIRECTORY = path.dirname(__file__)

DEFAULT_PARAMS_YAML = path.join(SCRIPT_DIRECTORY, "../operator/params.yaml")

log = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    datefmt="%Y-%m-%d %H:%M:%SZ",
)


def generate_markdown_table(input_params_yaml, output_markdown_file_path):
    try:
        with open(input_params_yaml, "r") as stream:
            params = yaml.safe_load(stream)["parameters"]

            writer = MarkdownTableWriter()
            writer.table_name = "Parameters"
            writer.styles = [
                Style(align="left", font_weight="bold"),
                Style(align="left"),
                Style(align="right"),
            ]
            writer.headers = ["Name", "Description", "Default"]

            writer.value_matrix = [list(d.values()) for d in params]

            if output_markdown_file_path:
                with open(output_markdown_file_path, "w") as file:
                    writer.stream = file
                    writer.write_table()
            else:
                print(writer.dumps())

            return 0
    except Exception as e:
        logging.error(f"Failed to generate Markdown for {input_params_yaml}", e)
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
    output_markdown_file_path = path.join(
        SCRIPT_DIRECTORY, args.output_markdown_file
    )

    return generate_markdown_table(
        args.input_params_yaml, output_markdown_file_path
    )


if __name__ == "__main__":
    sys.exit(main())

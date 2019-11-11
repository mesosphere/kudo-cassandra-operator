import argparse
import logging
import yaml
from pytablewriter import MarkdownTableWriter
from pytablewriter.style import Style

DEFAULT_PARAMS_YAML = '../operator/params.yaml'
DEFAULT_PARAMS_MD = 'PARAMETERS.md'


def generateMarkdownTable(params_yaml, params_md):
    try:
        with open(params_yaml, 'r') as stream:
            params = yaml.safe_load(stream)['parameters']

            writer = MarkdownTableWriter()
            writer.table_name = "Parameters"
            writer.styles = [
                Style(align="left", font_weight="bold"),
                Style(align="left"),
                Style(align="right"),
            ]
            writer.headers = ["Name", "Description", "Default"]

            writer.value_matrix = [list(d.values()) for d in params]

            with open(params_md, "w") as file:
                writer.stream = file
                writer.write_table()
    except Exception as e:
        logging.error("Failed to generate PARAMETERS Markdown Table", e)


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-params_yaml', help='Params YAML input file.', default=DEFAULT_PARAMS_YAML)
    parser.add_argument('-params_md', help='Params markdown output file.', default=DEFAULT_PARAMS_MD)
    args = parser.parse_args()

    generateMarkdownTable(args.params_yaml, args.params_md)

#!/usr/bin/env python3

"""
Generate log-event reference doc from source code
"""


import os
import re

from rstcloth import RstCloth


def rstescape(txt):
    """
    Escape RST formatting symbols from text
    """
    return txt.replace("*", r"\*")


def parse_structs(file_content):
    """
    Parse information about a particular struct in the generated file
    """
    pattern = re.compile(
        r"// BEGIN (\w+)\n\n"  # Match typename
        r"// auto-generated.*\n"  # Match warning
        r"//\s*([\w./]+)\.\1\s*\n"  # Match package
        r"((?:\s*//.*\n)+)"  # Match comments
        r"(type\s+\1\s+struct\s*{(?:.|\n)+?})\s*"  # Match typedef
        r"// END \1",  # Match end
        re.MULTILINE,
    )

    structs = []

    for match in pattern.finditer(file_content):
        typename = match.group(1)
        packagename = match.group(2)
        comments = match.group(3).replace("//", "").strip()
        typedef = match.group(4).strip()

        struct_info = {
            "typename": typename,
            "packagename": packagename,
            "fullcomment": comments,
            "typedef": typedef,
        }

        structs.append(struct_info)

    return structs


def output_section_common(common_modules, rst, h2_common, h3_common, text):
    """
    Output section with common modules
    """

    rst.h2(h2_common)

    sorted_modules = sorted(common_modules.keys())
    for module in sorted_modules:
        rst.h3(f"{h3_common} {module.replace('ivxv.ee/', '')}")

        sorted_events = sorted(common_modules[module].keys())
        for event in sorted_events:
            data = common_modules[module][event]
            rst.definition(event, text)
            _, val = next(iter(data.items()))
            rst.codeblock(val["typedef"], indent=3)
            rst.newline()
            for mod in data:
                rst.li(rstescape(f"{mod}.{event}: {data[mod]['fullcomment']}"))

            rst.newline()


def output_section_unique(unique_modules, rst, h2_unique, h3_unique):
    """
    Output section with unique models
    """

    rst.h2(h2_unique)
    sorted_modules = sorted(unique_modules.keys())
    for module in sorted_modules:
        rst.h3(f"{h3_unique} {module.replace('ivxv.ee/', '')}")

        sorted_events = sorted(unique_modules[module].keys())
        for event in sorted_events:
            data = unique_modules[module][event]
            rst.definition(
                f"{data['packagename']}.{event}", rstescape(data["fullcomment"])
            )
            rst.codeblock(data["typedef"], indent=3)
            rst.newline()


def output_intro(rst, h1, content_file):
    """
    Output introduction
    """

    rst.h1(h1)
    with open(content_file, "r", encoding="ascii") as inf:
        for line in inf.readlines():
            if line.startswith("*"):
                rst.li(line.replace("* ", ""))
            else:
                rst.content(line)
        rst.newline()


def categorize_event(pkg, evt, item, exclusions, events_dicts):
    """
    Analyze each event to a category
    """

    if exclusions and pkg in exclusions:
        return
    if "/cmd/" in pkg:
        events_dicts["cmd_events"].setdefault(pkg, {})[evt] = item
    elif "/service/" in pkg:
        events_dicts["srv_events"].setdefault(pkg, {})[evt] = item
    elif "/internal/" in pkg:
        events_dicts["int_events"].setdefault(pkg, {})[evt] = item
    else:
        events_dicts["events"].setdefault(pkg, {})[evt] = item


def analyze_src(src_dir, target_file, exclusions=None):
    """
    Walk the source tree and analyze generated code
    """

    events_dicts = {"events": {}, "cmd_events": {}, "int_events": {}, "srv_events": {}}

    for dirpath, _, filenames in os.walk(src_dir):
        if target_file in filenames:
            full_path = os.path.join(dirpath, target_file)

            with open(full_path, "r", encoding="utf8") as file:
                parsed_structs = parse_structs(file.read())

                for item in parsed_structs:
                    categorize_event(
                        item["packagename"],
                        item["typename"],
                        item,
                        exclusions,
                        events_dicts,
                    )

    return tuple(
        events_dicts[key]
        for key in ["events", "cmd_events", "int_events", "srv_events"]
    )


def remove_common(events):
    """
    Remove common events from unique set
    """

    common_events = {}
    unique_events = {}

    all_events = {}

    for module, mod_events in events.items():
        for event in mod_events:
            all_events.setdefault(event, set([]))
            all_events[event].add(module)

    for event, modules in all_events.items():
        if len(modules) > 1:
            by_types = {}
            for mod in modules:
                typedef = events[mod][event]["typedef"]
                by_types.setdefault(typedef, set([]))
                by_types[typedef].add(mod)

            if len(by_types) > 1:
                pass
                # This should not happen often. This means that same identifiers
                # do not have equal typedefs and we may even have several different
                # clusters with same identifier
                # TODO
                # print(event,file=sys.stderr)
                # print(by_types, file=sys.stderr)

            for _, val in by_types.items():
                if len(val) > 1:

                    printable = ""
                    for mod in sorted(val):
                        printable += mod
                        printable += ", "

                    printable = printable[:-2]

                    common_events.setdefault(printable, {})

                    common_events[printable][event] = {}
                    for mod in sorted(val):
                        common_events[printable][event][mod] = events[mod][event]

                else:
                    mod = next(iter(val))
                    unique_events.setdefault(mod, {})
                    unique_events[mod][event] = events[mod][event]
        else:
            mod = next(iter(modules))
            unique_events.setdefault(mod, {})
            unique_events[mod][event] = events[mod][event]

    return common_events, unique_events


if __name__ == "__main__":

    SRC_DIR = os.path.abspath("../../../")
    print(SRC_DIR)
    TARGET_FILE = "gen_types.go"

    OUT_FILE = "index.rst"

    EXCLUSIONS = set(
        [
            "ivxv.ee/common/collector/auth/dummy",
            "ivxv.ee/common/collector/container/dummy",
            "ivxv.ee/common/collector/storage/file",
            "ivxv.ee/common/collector/storage/memory",
        ]
    )

    MOD_EVENTS, CMD_EVENTS, INT_EVENTS, SRV_EVENTS = analyze_src(
        SRC_DIR, TARGET_FILE, EXCLUSIONS
    )

    with open(OUT_FILE, "w") as output_file:

        RST = RstCloth(output_file)
        RST.title("IVXV Backend Log Messages")

        RST.newline()
        RST.content(".. raw:: html")
        RST.newline()
        RST.content('   <p style="background-color: #f99; padding: 20px;">')
        RST.content("     <strong>NB!</strong>")
        RST.content("     See on HTML-versioon dokumendist.")
        RST.content("     Tellijale antakse üle PDF-versioon.")
        RST.content("   </p>")
        RST.newline()
        RST.content(".. toctree::")
        RST.content("   :maxdepth: 4")
        RST.newline()

        output_intro(RST, "Overview", "introduction.inc")

        CMD_COMMON, CMD_UNIQUE = remove_common(CMD_EVENTS)
        SRV_COMMON, SRV_UNIQUE = remove_common(SRV_EVENTS)
        MOD_COMMON, MOD_UNIQUE = remove_common(MOD_EVENTS)
        INT_COMMON, INT_UNIQUE = remove_common(INT_EVENTS)

        RST.h1("Commands")
        output_section_common(
            CMD_COMMON, RST, "Common events", "Commands", "Typedef for common event"
        )
        output_section_unique(CMD_UNIQUE, RST, "Specific events", "Command")

        RST.h1("Services")
        output_section_common(
            SRV_COMMON, RST, "Common events", "Services", "Typedef for common event"
        )
        output_section_unique(SRV_UNIQUE, RST, "Specific events", "Service")

        RST.h1("Modules")
        output_section_common(
            MOD_COMMON, RST, "Common events", "Modules", "Typedef for common event"
        )
        output_section_unique(MOD_UNIQUE, RST, "Specific events", "Module")

        RST.h1("Internal modules")
        output_section_common(
            INT_COMMON, RST, "Common events", "Modules", "Typedef for common event"
        )
        output_section_unique(INT_UNIQUE, RST, "Specific events", "Module")

#!/usr/bin/env bash

generate() {
    generateForNode
    generateForKeyGenerator
    generateForTermUi
    generateForLogViewer
    generateForSeedNode
}

generateForNode() {
    HELP="
# Node CLI

The **Dharitri Node** exposes the following Command Line Interface:
$(code)
\$ node --help

$(./node/node --help | head -n -3)
$(code)
"
    echo "$HELP" > ./node/CLI.md
}

generateForKeyGenerator() {
    HELP="
# Keygenerator CLI

The **Key generation Tool** exposes the following Command Line Interface:
$(code)
\$ keygenerator --help

$(./keygenerator/keygenerator --help | head -n -3)
$(code)
"
    echo "$HELP" > ./keygenerator/CLI.md
}

generateForTermUi() {
    HELP="
# Dharitri TermUI CLI

The **Dharitri Term UI** exposes the following Command Line Interface:
$(code)
\$ termui --help

$(./termui/termui --help | head -n -3)
$(code)
"
    echo "$HELP" > ./termui/CLI.md
}

generateForLogViewer() {
    HELP="
# Logviewer App

The **Dharitri Logviewer App** exposes the following Command Line Interface:
$(code)
\$ logviewer --help

$(./logviewer/logviewer --help | head -n -3)
$(code)
"
    echo "$HELP" > ./logviewer/CLI.md
}

generateForSeedNode() {
    HELP="
# Dharitri SeedNode CLI

The **Dharitri SeedNode** exposes the following Command Line Interface:
$(code)
\$ seednode --help

$(./seednode/seednode --help | head -n -3)
$(code)
"
    echo "$HELP" > ./seednode/CLI.md
}

code() {
    printf "\n\`\`\`\n"
}

generate

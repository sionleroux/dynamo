cat 1.txt | tail -n +2 | sed -e '/67/s/.*/0/' -e '/199/s/.*/1/' | tr '\n' ', ' > title.txt

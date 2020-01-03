
if [ ! -n "$1" ] ;then
        echo "Please inpu log file name!"
        exit
fi

if [ ! -w "$1" ];then
        echo "$1 not exist!"
        exit
fi

current=`date "+%Y-%m-%d %H:%M:%S"`
timeStamp=`date -d "$current" +%s`
now=$(((timeStamp*1000+10#`date "+%N"`/1000000)/1000))

grep monitor $1 > analysising_temp_$now
awk -F[=,\(] '{print $2,$4,$6,$8,$10}' analysising_temp_$now
rm -rf analysising_temp_$now


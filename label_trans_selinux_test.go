// +build selinux,linux

package setrans

import (
	"strings"
	"testing"
)

func check(t *testing.T, err error, wanterr, got, want string) {
	t.Helper()

	if err != nil {
		if wanterr == "" {
			t.Fatalf("got error %q when no error expected", err)
		}
		if !strings.Contains(err.Error(), wanterr) {
			t.Fatalf("got error %q does not match expected err %q", err, wanterr)
		}
		return
	}

	if got != want {
		t.Errorf("translated %q does not match expected %q", got, want)
	}
}

// These tests require an MLS policy and will not work with the default
// Redhat targeted policy
func TestTranslation(t *testing.T) {
	t.Run("Test TransToRaw", func(t *testing.T) {
		tests := []struct {
			orig, want, err, name string
		}{
			{
				orig: "dbclient_u:dbclient_r:dbclient_t:T REL TO USA",
				want: "dbclient_u:dbclient_r:dbclient_t:s6:c0,c2,c11,c201.c429,c431.c511",
				name: "test a valid country",
			},
			{
				orig: "dbclient_u:dbclient_r:dbclient_t:T REL TO NATO",
				want: "dbclient_u:dbclient_r:dbclient_t:s6:c0,c2,c11,c201.c204,c206.c218,c220.c222,c224.c238,c240.c256,c259,c260,c262.c267,c270.c273,c275.c277,c279.c287,c289.c297,c299,c301.c307,c309,c311.c330,c334.c364,c367.c377,c379,c380,c382.c386,c388.c405,c408.c422,c424.c429,c431.c511",
				name: "test a group",
			},
			{
				orig: "dbclient_u:dbclient_r:dbclient_t:C REL TO USA,ABW,ATG,ARE,AFG,DZA,AZE,ALB,ARM,AND,AGO,ASM,ARG,AUS,AUT,AIA,ATA,BHR,BRB,BWA,BMU,BEL,BHS,BGD,BLZ,BIH,BOL,MMR,BEN,BLR,SLB,BRA,BTN,BGR,BVT,BRN,BDI,CAN,KHM,TCD,LKA,COG,COD,CHN,CHL,CYM,CCK,CMR,COM,COL,MNP,CRI,CAF,CUB,CPV,COK,CYP,DNK,DJI,DMA,DOM,ECU,EGY,IRL,GNQ,EST,ERI,SLV,ETH,CZE,GUF,FIN,FJI,FLK,FSM,FRO,PYF,FRA,ATF,GMB,GAB,GEO,GHA,GIB,GRD,GGY,GRL,DEU,GLP,GUM,GRC,GTM,GIN,GUY,PSE,HTI,HKG,HMD,HND,HRV,HUN,ISL,IDN,IMN,IND,IOT,IRN,ISR,ITA,CIV,IRQ,JPN,JEY,JAM,JOR,KEN,KGZ,PRK,KIR,ROK,CXR,KWT,KAZ,LAO,LBN,LVA,LTU,LBR,SVK,LIE,LSO,LUX,LBY,MDG,MTQ,MAC,MDA,MYT,MNG,MSR,MWI,MNE,MKD,MLI,MCO,MAR,MUS,MRT,MLT,OMN,MDV,MEX,MYS,MOZ,NCL,NIU,NFK,NER,VUT,NGA,NLD,NOR,NPL,NRU,SUR,ANT,NIC,NZL,PRY,PCN,PER,PAK,POL,PAN,PRT,PNG,PLW,GNB,QAT,SRB,REU,MHL,MAF,ROU,PHL,PRI,RUS,RWA,SAU,SPM,KNA,SYC,ZAF,SEN,SHN,SVN,SLE,SMR,SGP,SOM,ESP,LCA,SDN,SJM,SWE,SGS,SYR,CHE,BLM,TTO,THA,TJK,TCA,TGO,TKL,TON,STP,TUN,TLS,TUR,TUV,TWN,TKM,TZA,UGA,GBR,UKR,BFA,URY,UZB,VCT,VEN,VGB,VNM,VIR,VAT,NAM,WLF,ESH,WSM,SWZ,YEM,ZMB,ZWE",
				want: "dbclient_u:dbclient_r:dbclient_t:s4:c0,c2,c11,c445.c511",
				name: "test a long translation",
			},
			{
				orig: "staff_u:staff_r:staff_t:FooLow-FooHigh",
				want: "staff_u:staff_r:staff_t:FooLow-FooHigh",
				name: "invalid original context to TransToRaw",
			},
		}

		conn, err := New()
		if err != nil {
			t.Fatalf("failed to connect to mcstrans: %q", err)
		}
		defer conn.Close()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := conn.TransToRaw(tt.orig)
				check(t, err, tt.err, got, tt.want)
			})
		}
	})

	t.Run("Test RawToColor", func(t *testing.T) {
		tests := []struct {
			orig, want, err, name string
		}{
			{
				orig: "staff_u:staff_r:staff_t:s0-s15:c0.c1023",
				want: "#000000 #ffffff #000000 #ffffff #000000 #ffffff #000000 #ffffff",
				name: "test a valid RawToColor",
			},
		}

		conn, err := New()
		if err != nil {
			t.Fatalf("failed to connect to mcstrans: %q", err)
		}
		defer conn.Close()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := conn.RawToColor(tt.orig)
				check(t, err, tt.err, got, tt.want)
			})
		}
	})

	t.Run("Test RawToTrans", func(t *testing.T) {
		tests := []struct {
			orig, want, err, name string
		}{
			{
				orig: "staff_u:staff_r:staff_t:s0-s15:c0.c1023",
				want: "staff_u:staff_r:staff_t:SystemLow-SystemHigh",
				name: "test a valid RawToTrans",
			},
			{
				orig: "staff_u:staff_r:staff_t:s5:c0,c2,c11,c201.c214,c216.c238,c240.c277,c279.c368,c370.c429,c431.c511",
				want: "staff_u:staff_r:staff_t:S REL TO FVEY",
				name: "test a group",
			},
		}

		conn, err := New()
		if err != nil {
			t.Fatalf("failed to connect to mcstrans: %q", err)
		}
		defer conn.Close()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := conn.RawToTrans(tt.orig)
				check(t, err, tt.err, got, tt.want)
			})
		}
	})
}

func BenchmarkTranslation(b *testing.B) {
	input := "staff_u:staff_r:staff_t:SystemLow-SystemHigh"

	conn, err := New()
	if err != nil {
		b.Fatalf("failed to connect to mcstrans: %q", err)
	}
	defer conn.Close()

	for n := 0; n < b.N; n++ {
		if _, err := conn.TransToRaw(input); err != nil {
			b.Fatalf("failed to translate to raw: %v", err)
		}
	}
}

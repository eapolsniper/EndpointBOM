#!/usr/bin/env python3
"""
Dependency-Track SBOM Upload Script for Developer Workstations

This script uploads multiple SBOM files (package managers, applications, 
IDE extensions, browser extensions) to Dependency-Track, organizing them 
under a parent project by hostname.

Features:
- Project hierarchy (parent by hostname, children by SBOM type)
- Timestamp-based versioning for change tracking
- Classifiers to categorize project types
- Custom properties with endpoint metadata (user, IPs)
- Automatic BOM processing monitoring
"""

import requests
import json
import base64
import sys
import time
from pathlib import Path
from typing import Optional, Dict, List
from datetime import datetime, timedelta

# Configuration - UPDATE THESE FOR YOUR ENVIRONMENT
DEPENDENCY_TRACK_URL = "http://localhost:8081"  # Change to your Dependency-Track URL
API_KEY = "odt_YOUR_API_KEY_HERE"  # Get from Dependency-Track: Settings → Teams → API Keys

# SBOM file mapping with classifiers
# Classifiers: APPLICATION, FRAMEWORK, LIBRARY, CONTAINER, OPERATING_SYSTEM, DEVICE, FIRMWARE, FILE
SBOM_TYPES = {
    "package-managers": {
        "classifier": "LIBRARY",
        "description": "Dependencies from package managers (npm, pip, etc.)"
    },
    "applications": {
        "classifier": "APPLICATION", 
        "description": "Installed desktop applications"
    },
    "ide-extensions": {
        "classifier": "LIBRARY",
        "description": "IDE extensions and MCP servers"
    },
    "browser-extensions": {
        "classifier": "LIBRARY",
        "description": "Browser extensions and plugins"
    }
}



class DependencyTrackClient:
    """Client for interacting with Dependency-Track API"""
    
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.headers = {
            "X-Api-Key": api_key,
            "Content-Type": "application/json"
        }
    
    def create_project(self, 
                      name: str, 
                      version: str,
                      classifier: str,
                      description: str = "",
                      parent_uuid: Optional[str] = None,
                      tags: Optional[List[Dict[str, str]]] = None,
                      active: bool = True,
                      auto_create: bool = True) -> Optional[Dict]:
        """
        Create a new project in Dependency-Track
        
        API: PUT /api/v1/project
        
        Parameters:
        - name: Project name
        - version: Project version
        - classifier: PROJECT classifier (APPLICATION, FRAMEWORK, LIBRARY, etc.)
        - description: Project description
        - parent_uuid: UUID of parent project (for project hierarchy)
        - tags: List of tag dictionaries with 'name' and optional 'value'
        - active: Whether project is active
        - auto_create: Whether to auto-create if doesn't exist
        
        Returns:
        - Project object with UUID
        """
        print(f"\n📦 Creating project: {name} v{version}")
        print(f"   Classifier: {classifier}")
        if parent_uuid:
            print(f"   Parent UUID: {parent_uuid}")
        
        payload = {
            "name": name,
            "version": version,
            "classifier": classifier,
            "description": description,
            "active": active
        }
        
        if parent_uuid:
            payload["parent"] = {"uuid": parent_uuid}
        
        if tags:
            payload["tags"] = tags
        
        response = requests.put(
            f"{self.base_url}/api/v1/project",
            headers=self.headers,
            json=payload
        )
        
        if response.status_code == 201:
            project = response.json()
            print(f"   ✅ Created project UUID: {project['uuid']}")
            return project
        else:
            print(f"   ❌ Failed to create project: {response.status_code}")
            print(f"      {response.text}")
            return None
    
    def get_project(self, name: str, version: str) -> Optional[Dict]:
        """
        Get an existing project by name and version
        
        API: GET /api/v1/project/lookup?name={name}&version={version}
        Note: requests library automatically URL-encodes params
        """
        response = requests.get(
            f"{self.base_url}/api/v1/project/lookup",
            headers=self.headers,
            params={"name": name, "version": version}
        )
        
        if response.status_code == 200:
            return response.json()
        elif response.status_code == 404:
            return None
        else:
            print(f"   ⚠️  Error looking up project: {response.status_code}")
            return None
    
    def update_project_properties(self, 
                                  project_uuid: str,
                                  properties: List[Dict[str, str]]) -> bool:
        """
        Add custom properties to a project
        
        API: PUT /api/v1/project/{uuid}/property
        
        Properties format:
        [{
            "groupName": "Custom",
            "propertyName": "location", 
            "propertyValue": "Chicago Office",
            "propertyType": "STRING",
            "description": "Physical location"
        }]
        """
        print(f"\n🏷️  Adding properties to project {project_uuid}")
        
        for prop in properties:
            response = requests.put(
                f"{self.base_url}/api/v1/project/{project_uuid}/property",
                headers=self.headers,
                json=prop
            )
            
            if response.status_code == 201:
                print(f"   ✅ Added property: {prop['propertyName']}")
            else:
                print(f"   ❌ Failed to add property: {response.status_code}")
        
        return True
    
    def upload_bom(self, 
                  project_uuid: str, 
                  bom_file: Path,
                  auto_create: bool = False) -> Optional[str]:
        """
        Upload a BOM file to a project
        
        API: PUT /api/v1/bom
        
        Parameters:
        - project_uuid: UUID of the project
        - bom_file: Path to BOM file
        - auto_create: Auto-create project if it doesn't exist
        
        Returns:
        - Token for tracking BOM processing status
        """
        print(f"\n📤 Uploading BOM: {bom_file.name}")
        print(f"   Project UUID: {project_uuid}")
        
        with open(bom_file, 'r') as f:
            bom_content = f.read()
        
        # Encode BOM as base64
        bom_base64 = base64.b64encode(bom_content.encode()).decode()
        
        payload = {
            "project": project_uuid,
            "bom": bom_base64,
            "autoCreate": auto_create
        }
        
        response = requests.put(
            f"{self.base_url}/api/v1/bom",
            headers=self.headers,
            json=payload
        )
        
        if response.status_code == 200:
            result = response.json()
            token = result.get("token")
            print(f"   ✅ BOM uploaded successfully")
            print(f"   📋 Processing token: {token}")
            return token
        else:
            print(f"   ❌ Failed to upload BOM: {response.status_code}")
            print(f"      {response.text}")
            return None
    
    def check_bom_processing_status(self, token: str) -> Dict:
        """
        Check the processing status of an uploaded BOM
        
        API: GET /api/v1/bom/token/{token}
        
        Returns status: PROCESSING, PROCESSED, or FAILED
        """
        response = requests.get(
            f"{self.base_url}/api/v1/bom/token/{token}",
            headers=self.headers
        )
        
        if response.status_code == 200:
            return response.json()
        else:
            return {"processing": False, "status": "UNKNOWN"}
    


def extract_hostname_from_bom(bom_file: Path) -> Optional[str]:
    """Extract hostname from BOM metadata"""
    try:
        with open(bom_file, 'r') as f:
            bom = json.load(f)
        return bom.get("metadata", {}).get("component", {}).get("name")
    except:
        return None


def extract_metadata_from_bom(bom_file: Path) -> Dict:
    """Extract useful metadata from BOM file"""
    try:
        with open(bom_file, 'r') as f:
            bom = json.load(f)
        
        metadata = bom.get("metadata", {})
        component = metadata.get("component", {})
        properties = {prop["name"]: prop["value"] 
                     for prop in component.get("properties", [])}
        
        # Extract all local IPs (there can be multiple)
        local_ips = [prop["value"] for prop in component.get("properties", []) 
                    if prop.get("name") == "local_ip"]
        
        return {
            "hostname": component.get("name", "unknown"),
            "os": properties.get("os", "unknown"),
            "os_version": properties.get("os_version", "unknown"),
            "scan_category": properties.get("scan_category", "unknown"),
            "logged_in_user": properties.get("logged_in_user", ""),
            "local_ips": local_ips,
            "public_ip": properties.get("public_ip", ""),
            "timestamp": metadata.get("timestamp", "unknown"),
            "component_count": len(bom.get("components", []))
        }
    except Exception as e:
        print(f"Error extracting metadata: {e}")
        return {}


def format_version_from_timestamp(timestamp_str: str) -> str:
    """
    Convert BOM timestamp to version format (YYYY-MM-DD-HHMM)
    Example: "2025-12-13T16:54:43-06:00" -> "2025-12-13-1654"
    """
    try:
        from dateutil import parser
        dt = parser.parse(timestamp_str)
        return dt.strftime("%Y-%m-%d-%H%M")
    except:
        # Fallback: try simple parsing
        try:
            # Handle ISO format: 2025-12-13T16:54:43-06:00
            import re
            match = re.match(r'(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):\d{2}', timestamp_str)
            if match:
                year, month, day, hour, minute = match.groups()
                return f"{year}-{month}-{day}-{hour}{minute}"
        except:
            pass
        # Final fallback: use current time
        return datetime.now().strftime("%Y-%m-%d-%H%M")


def build_project_description(base_description: str, metadata: Dict) -> str:
    """
    Build project description including optional custom data from SBOM metadata.
    Includes: logged_in_user, local IPs, public IP
    """
    description_parts = [base_description]
    
    # Add logged in user if available
    if metadata.get("logged_in_user"):
        description_parts.append(f"User: {metadata['logged_in_user']}")
    
    # Add local IPs if available
    local_ips = metadata.get("local_ips", [])
    if local_ips:
        ips_str = ", ".join(local_ips)
        description_parts.append(f"Local IPs: {ips_str}")
    
    # Add public IP if available
    if metadata.get("public_ip"):
        description_parts.append(f"Public IP: {metadata['public_ip']}")
    
    return " | ".join(description_parts)


def get_most_recent_scan_files(scans_dir: Path) -> List[Path]:
    """
    Find the most recent scan files by timestamp in the filename.
    Expected format: hostname.timestamp-TZ.type.cdx.json
    Note: hostname may contain dots (e.g., "host.local")
    Returns only files from the most recent scan.
    """
    all_files = list(scans_dir.glob("*.cdx.json"))
    if not all_files:
        return []
    
    # Extract timestamps from filenames (format: hostname.YYYYMMDD-HHMMSS-TZ.type.cdx.json)
    file_timestamps = {}
    for file in all_files:
        parts = file.stem.split('.')
        
        # Find the timestamp part (format: YYYYMMDD-HHMMSS-TZ or YYYYMMDD-HHMMSS for backward compatibility)
        timestamp_str = None
        for part in parts:
            # Check if this part looks like a timestamp (with or without timezone)
            if len(part) >= 15 and part[8] == '-' and part[:8].isdigit() and part[9:15].isdigit():
                # Extract just the date-time part (without timezone for parsing)
                if len(part) > 15 and part[15] == '-':
                    timestamp_str = part[:15]  # Strip timezone
                else:
                    timestamp_str = part
                break
        
        if not timestamp_str:
            print(f"⚠️  Skipping file with invalid timestamp format: {file.name}")
            continue
        
        try:
            # Parse timestamp to validate format
            timestamp = datetime.strptime(timestamp_str, "%Y%m%d-%H%M%S")
            file_timestamps[file] = timestamp
        except ValueError:
            print(f"⚠️  Skipping file with invalid timestamp format: {file.name}")
            continue
    
    if not file_timestamps:
        return []
    
    # Find the most recent timestamp
    most_recent_timestamp = max(file_timestamps.values())
    
    # Get all files matching the most recent timestamp
    recent_files = [f for f, ts in file_timestamps.items() if ts == most_recent_timestamp]
    
    return recent_files


def archive_files(files: List[Path], scans_dir: Path) -> None:
    """Move uploaded files to archive subdirectory"""
    archive_dir = scans_dir / "archive"
    archive_dir.mkdir(exist_ok=True)
    
    print(f"\n📦 Archiving {len(files)} uploaded files...")
    for file in files:
        dest = archive_dir / file.name
        file.rename(dest)
        print(f"   ✅ Moved to archive: {file.name}")


def cleanup_old_files(scans_dir: Path, days: int = 60) -> None:
    """Remove files older than specified days from scans/ and scans/archive/"""
    cutoff_date = datetime.now() - timedelta(days=days)
    
    print(f"\n🧹 Cleaning up files older than {days} days...")
    
    # Check both main scans directory and archive
    directories = [scans_dir, scans_dir / "archive"]
    
    total_removed = 0
    for directory in directories:
        if not directory.exists():
            continue
        
        for file in directory.glob("*.cdx.json"):
            # Get file modification time
            file_mtime = datetime.fromtimestamp(file.stat().st_mtime)
            
            if file_mtime < cutoff_date:
                file.unlink()
                total_removed += 1
                print(f"   🗑️  Removed old file: {file.relative_to(scans_dir.parent)}")
    
    if total_removed == 0:
        print("   ✅ No old files to remove")
    else:
        print(f"   ✅ Removed {total_removed} old file(s)")


def main():
    """Main execution function"""
    print("=" * 80)
    print("Dependency-Track SBOM Upload Script")
    print("=" * 80)
    
    # Initialize client
    client = DependencyTrackClient(DEPENDENCY_TRACK_URL, API_KEY)
    
    # Find SBOM files in scans directory
    scans_dir = Path("scans")
    if not scans_dir.exists():
        print(f"❌ Scans directory not found: {scans_dir}")
        sys.exit(1)
    
    # Get only the most recent scan files
    print(f"\n🔍 Looking for most recent scan files in {scans_dir}...")
    sbom_files = get_most_recent_scan_files(scans_dir)
    
    if not sbom_files:
        print(f"❌ No valid SBOM files found in {scans_dir}")
        sys.exit(1)
    
    # Extract timestamp from first file for logging
    first_file_parts = sbom_files[0].stem.split('.')
    scan_timestamp = first_file_parts[1] if len(first_file_parts) >= 2 else "unknown"
    
    print(f"\n📁 Found most recent scan: {scan_timestamp}")
    print(f"   Files to upload ({len(sbom_files)}):")
    for sbom in sbom_files:
        print(f"   - {sbom.name}")
    
    # Group files by hostname
    hostname_groups = {}
    for sbom_file in sbom_files:
        hostname = extract_hostname_from_bom(sbom_file)
        if hostname:
            if hostname not in hostname_groups:
                hostname_groups[hostname] = []
            hostname_groups[hostname].append(sbom_file)
    
    print(f"\n🖥️  Found {len(hostname_groups)} unique hostname(s):")
    for hostname in hostname_groups:
        print(f"   - {hostname} ({len(hostname_groups[hostname])} BOMs)")
    
    # Process each hostname
    for hostname, bom_files in hostname_groups.items():
        print("\n" + "=" * 80)
        print(f"Processing hostname: {hostname}")
        print("=" * 80)
        
        # Get metadata from first BOM
        first_bom = bom_files[0]
        metadata = extract_metadata_from_bom(first_bom)
        
        print(f"\n📊 Endpoint Metadata:")
        print(f"   Hostname: {metadata.get('hostname', 'unknown')}")
        print(f"   OS: {metadata.get('os', 'unknown')} {metadata.get('os_version', 'unknown')}")
        print(f"   User: {metadata.get('logged_in_user', 'unknown')}")
        print(f"   Scan Time: {metadata.get('timestamp', 'unknown')}")
        
        # Step 1: Create parent project for the hostname
        print("\n" + "-" * 80)
        print("STEP 1: CREATE PARENT PROJECT")
        print("-" * 80)
        
        parent_name = hostname
        # Use static "latest" version for parent - children will have timestamped versions
        parent_version = "latest"
        print(f"   Parent version: {parent_version} (children will use scan timestamp)")
        
        # Check if parent already exists
        parent_project = client.get_project(parent_name, parent_version)
        
        if not parent_project:
            # Create parent project as DEVICE type
            parent_project = client.create_project(
                name=parent_name,
                version=parent_version,
                classifier="DEVICE",
                description=f"Developer workstation: {hostname}"
            )
        else:
            print(f"\n📦 Parent project already exists: {parent_project['uuid']}")
        
        if not parent_project:
            print("❌ Failed to create parent project. Skipping hostname.")
            continue
        
        parent_uuid = parent_project['uuid']
        
        # Add custom properties to parent project
        custom_properties = [
            {
                "groupName": "Endpoint Information",
                "propertyName": "operating_system",
                "propertyValue": metadata.get('os', 'unknown'),
                "propertyType": "STRING",
                "description": "Operating system type"
            },
            {
                "groupName": "Endpoint Information",
                "propertyName": "os_version",
                "propertyValue": metadata.get('os_version', 'unknown'),
                "propertyType": "STRING",
                "description": "Operating system version"
            },
            {
                "groupName": "Endpoint Information",
                "propertyName": "logged_in_user",
                "propertyValue": metadata.get('logged_in_user', 'unknown'),
                "propertyType": "STRING",
                "description": "Primary logged-in user"
            },
            {
                "groupName": "Compliance",
                "propertyName": "scan_frequency",
                "propertyValue": "daily",
                "propertyType": "STRING",
                "description": "How often this endpoint is scanned"
            },
            {
                "groupName": "Compliance",
                "propertyName": "last_scan_time",
                "propertyValue": metadata.get('timestamp', 'unknown'),
                "propertyType": "STRING",
                "description": "Timestamp of last scan"
            }
        ]
        
        client.update_project_properties(parent_uuid, custom_properties)
        
        # Step 2: Create child projects and upload BOMs
        print("\n" + "-" * 80)
        print("STEP 2: CREATE CHILD PROJECTS & UPLOAD BOMs")
        print("-" * 80)
        
        upload_tokens = []
        
        for bom_file in bom_files:
            # Determine SBOM type from filename
            sbom_type = None
            for type_key in SBOM_TYPES.keys():
                if type_key in bom_file.name:
                    sbom_type = type_key
                    break
            
            if not sbom_type:
                print(f"\n⚠️  Could not determine SBOM type for {bom_file.name}, skipping")
                continue
            
            type_config = SBOM_TYPES[sbom_type]
            bom_metadata = extract_metadata_from_bom(bom_file)
            
            # Create child project with timestamp-based version
            child_name = f"{hostname} - {sbom_type}"
            # Use timestamp from BOM for child version
            child_version = format_version_from_timestamp(bom_metadata.get('timestamp', metadata.get('timestamp', '')))
            
            print(f"\n   Creating version: {child_version} for {sbom_type}")
            
            # Build description with custom data from this specific SBOM
            child_description = build_project_description(
                type_config["description"],
                bom_metadata
            )
            
            # Check if child already exists
            child_project = client.get_project(child_name, child_version)
            
            if not child_project:
                child_project = client.create_project(
                    name=child_name,
                    version=child_version,
                    classifier=type_config["classifier"],
                    description=child_description,
                    parent_uuid=parent_uuid
                )
            else:
                print(f"\n📦 Child project already exists: {child_project['uuid']}")
            
            if not child_project:
                print(f"❌ Failed to create child project for {sbom_type}")
                continue
            
            child_uuid = child_project['uuid']
            
            # Upload BOM to child project
            token = client.upload_bom(child_uuid, bom_file)
            if token:
                upload_tokens.append({
                    "token": token,
                    "project": child_name,
                    "type": sbom_type
                })
        
        # Step 3: Monitor processing status
        print("\n" + "-" * 80)
        print("STEP 3: MONITORING BOM PROCESSING")
        print("-" * 80)
        
        print(f"\n⏳ Waiting for {len(upload_tokens)} BOMs to process...")
        print("   (This may take a few moments)\n")
        
        for item in upload_tokens:
            max_attempts = 30
            attempt = 0
            while attempt < max_attempts:
                status = client.check_bom_processing_status(item['token'])
                if not status.get('processing', True):
                    print(f"   ✅ {item['project']} - Processing complete")
                    break
                attempt += 1
                time.sleep(2)
            
            if attempt >= max_attempts:
                print(f"   ⏰ {item['project']} - Still processing (timeout)")
        
        # Step 4: Display summary
        print("\n" + "-" * 80)
        print("STEP 4: UPLOAD SUMMARY")
        print("-" * 80)
        
        print(f"\n✅ Successfully uploaded {len(upload_tokens)} BOMs for {hostname}")
        print(f"\n🔗 View in Dependency-Track:")
        print(f"   http://localhost:8080/projects/{parent_uuid}")
        
        print(f"\n📊 Project Structure:")
        print(f"   Parent: {parent_name} (DEVICE)")
        for item in upload_tokens:
            print(f"      └─ {item['project']} ({SBOM_TYPES[item['type']]['classifier']})")
    
    print("\n" + "=" * 80)
    print("✅ All uploads complete!")
    print("=" * 80)
    
    # Archive uploaded files
    archive_files(sbom_files, scans_dir)
    
    # Cleanup old files (>60 days)
    cleanup_old_files(scans_dir, days=60)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n❌ Upload cancelled by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

